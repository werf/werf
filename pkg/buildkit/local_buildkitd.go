package buildkit

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"maps"
	"slices"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/logboek"
)

const (
	localBuildkitdContainerName   = "werf-buildkitd"
	localBuildkitdImage           = "moby/buildkit:v0.29.0"
	localBuildkitdConfigHashLabel = "werf.io/buildkitd-config-hash"
)

type ResolveHostOptions struct {
	// Registry addresses to configure as insecure in the werf-managed buildkitd
	// (no effect when an external buildkitd endpoint is set via environment).
	InsecureRegistryAddresses      []string
	SkipTLSVerifyRegistryAddresses []string
}

// ResolveHost returns the buildkitd endpoint to use: $WERF_BUILDKIT_HOST or $BUILDKIT_HOST
// when set, otherwise a werf-managed buildkitd container on the local Docker daemon.
func ResolveHost(ctx context.Context, opts ResolveHostOptions) (string, error) {
	if host := HostFromEnv(); host != "" {
		logboek.Context(ctx).Default().LogF("Using buildkit backend with buildkitd at %s\n", host)
		return host, nil
	}

	buildkitdConfig, err := makeLocalBuildkitdConfig(opts)
	if err != nil {
		return "", err
	}

	if err := ensureLocalBuildkitd(ctx, buildkitdConfig); err != nil {
		return "", fmt.Errorf("unable to set up local buildkitd container (alternatively set $WERF_BUILDKIT_HOST or $BUILDKIT_HOST to an external buildkitd endpoint): %w", err)
	}

	logboek.Context(ctx).Default().LogF("Using buildkit backend with local buildkitd container %q\n", localBuildkitdContainerName)
	return "docker-container://" + localBuildkitdContainerName, nil
}

func makeLocalBuildkitdConfig(opts ResolveHostOptions) (string, error) {
	registryOptions := map[string]map[string]bool{}
	setOption := func(addresses []string, key string) error {
		for _, address := range addresses {
			ref, err := name.ParseReference(address, name.WeakValidation)
			if err != nil {
				return fmt.Errorf("parse registry address %q: %w", address, err)
			}
			host := ref.Context().RegistryStr()
			if registryOptions[host] == nil {
				registryOptions[host] = map[string]bool{}
			}
			registryOptions[host][key] = true
		}
		return nil
	}

	if err := setOption(opts.InsecureRegistryAddresses, "http"); err != nil {
		return "", err
	}
	if err := setOption(opts.SkipTLSVerifyRegistryAddresses, "insecure"); err != nil {
		return "", err
	}

	var config bytes.Buffer
	for _, host := range slices.Sorted(maps.Keys(registryOptions)) {
		fmt.Fprintf(&config, "[registry.%q]\n", host)
		for _, key := range slices.Sorted(maps.Keys(registryOptions[host])) {
			fmt.Fprintf(&config, "  %s = true\n", key)
		}
	}
	return config.String(), nil
}

func localBuildkitdConfigHash(buildkitdConfig string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(localBuildkitdImage+"\n"+buildkitdConfig)))
}

func ensureLocalBuildkitd(ctx context.Context, buildkitdConfig string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		return fmt.Errorf("ping docker daemon: %w", err)
	}

	configHash := localBuildkitdConfigHash(buildkitdConfig)

	inspect, err := cli.ContainerInspect(ctx, localBuildkitdContainerName)
	switch {
	case err == nil && needRecreateLocalBuildkitd(inspect.Config, buildkitdConfig, configHash):
		logboek.Context(ctx).Default().LogF("Recreating local buildkitd container %q (configuration changed)\n", localBuildkitdContainerName)
		if err := cli.ContainerRemove(ctx, localBuildkitdContainerName, container.RemoveOptions{Force: true}); err != nil {
			return fmt.Errorf("remove container %q: %w", localBuildkitdContainerName, err)
		}
		if err := createLocalBuildkitdContainer(ctx, cli, buildkitdConfig, configHash); err != nil {
			return err
		}
	case err == nil:
		if inspect.State != nil && inspect.State.Running {
			return nil
		}
	case client.IsErrNotFound(err):
		if err := createLocalBuildkitdContainer(ctx, cli, buildkitdConfig, configHash); err != nil {
			return err
		}
	default:
		return fmt.Errorf("inspect container %q: %w", localBuildkitdContainerName, err)
	}

	if err := cli.ContainerStart(ctx, localBuildkitdContainerName, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container %q: %w", localBuildkitdContainerName, err)
	}

	return waitLocalBuildkitdReady(ctx)
}

// needRecreateLocalBuildkitd reports whether the existing container must be replaced to apply
// the desired registry configuration. Commands that need no registry configuration (empty
// desired config, e.g. background host cleanup) reuse any existing container as is, otherwise
// concurrent werf commands would recreate the container back and forth.
// ponytail: alternating builds with different insecure registries still recreate each time;
// merge configs if that becomes a real workflow.
func needRecreateLocalBuildkitd(containerConfig *container.Config, buildkitdConfig, configHash string) bool {
	if buildkitdConfig == "" {
		return false
	}
	return containerConfig == nil || containerConfig.Labels[localBuildkitdConfigHashLabel] != configHash
}

func createLocalBuildkitdContainer(ctx context.Context, cli *client.Client, buildkitdConfig, configHash string) error {
	if _, err := cli.ImageInspect(ctx, localBuildkitdImage); err != nil {
		logboek.Context(ctx).Default().LogF("Pulling %s\n", localBuildkitdImage)
		rc, err := cli.ImagePull(ctx, localBuildkitdImage, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("pull image %q: %w", localBuildkitdImage, err)
		}
		_, copyErr := io.Copy(io.Discard, rc)
		rc.Close()
		if copyErr != nil {
			return fmt.Errorf("pull image %q: %w", localBuildkitdImage, copyErr)
		}
	}

	logboek.Context(ctx).Default().LogF("Starting local buildkitd container %q\n", localBuildkitdContainerName)
	_, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:  localBuildkitdImage,
			Labels: map[string]string{localBuildkitdConfigHashLabel: configHash},
		},
		&container.HostConfig{
			Privileged:    true,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped},
		},
		nil, nil, localBuildkitdContainerName,
	)
	// Conflict: a concurrent werf process created the container first.
	if err != nil {
		if cerrdefs.IsConflict(err) {
			return nil
		}
		return fmt.Errorf("create container %q: %w", localBuildkitdContainerName, err)
	}

	if buildkitdConfig != "" {
		if err := copyLocalBuildkitdConfig(ctx, cli, buildkitdConfig); err != nil {
			return err
		}
	}
	return nil
}

func copyLocalBuildkitdConfig(ctx context.Context, cli *client.Client, buildkitdConfig string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "buildkit/buildkitd.toml", Mode: 0o644, Size: int64(len(buildkitdConfig))}); err != nil {
		return fmt.Errorf("write buildkitd config tar header: %w", err)
	}
	if _, err := tw.Write([]byte(buildkitdConfig)); err != nil {
		return fmt.Errorf("write buildkitd config tar data: %w", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("close buildkitd config tar: %w", err)
	}

	if err := cli.CopyToContainer(ctx, localBuildkitdContainerName, "/etc/", &buf, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("copy buildkitd config into container %q: %w", localBuildkitdContainerName, err)
	}
	return nil
}

func waitLocalBuildkitdReady(ctx context.Context) error {
	cl, err := NewClient(ctx, "docker-container://"+localBuildkitdContainerName)
	if err != nil {
		return err
	}
	defer cl.Close()

	deadline := time.Now().Add(30 * time.Second)
	for {
		listCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_, err := cl.ListWorkers(listCtx)
		cancel()
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("wait for buildkitd readiness in container %q: %w", localBuildkitdContainerName, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
