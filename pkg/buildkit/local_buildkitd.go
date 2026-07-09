package buildkit

import (
	"context"
	"fmt"
	"io"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	"github.com/werf/logboek"
)

const (
	localBuildkitdContainerName = "werf-buildkitd"
	localBuildkitdImage         = "moby/buildkit:v0.29.0"
)

// ResolveHost returns the buildkitd endpoint to use: $WERF_BUILDKIT_HOST or $BUILDKIT_HOST
// when set, otherwise a werf-managed buildkitd container on the local Docker daemon.
func ResolveHost(ctx context.Context) (string, error) {
	if host := HostFromEnv(); host != "" {
		logboek.Context(ctx).Default().LogF("Using buildkit backend with buildkitd at %s\n", host)
		return host, nil
	}

	if err := ensureLocalBuildkitd(ctx); err != nil {
		return "", fmt.Errorf("unable to set up local buildkitd container (alternatively set $WERF_BUILDKIT_HOST or $BUILDKIT_HOST to an external buildkitd endpoint): %w", err)
	}

	logboek.Context(ctx).Default().LogF("Using buildkit backend with local buildkitd container %q\n", localBuildkitdContainerName)
	return "docker-container://" + localBuildkitdContainerName, nil
}

func ensureLocalBuildkitd(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		return fmt.Errorf("ping docker daemon: %w", err)
	}

	inspect, err := cli.ContainerInspect(ctx, localBuildkitdContainerName)
	switch {
	case err == nil && inspect.State != nil && inspect.State.Running:
		return nil
	case err == nil:
	case client.IsErrNotFound(err):
		if err := createLocalBuildkitdContainer(ctx, cli); err != nil {
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

func createLocalBuildkitdContainer(ctx context.Context, cli *client.Client) error {
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
		&container.Config{Image: localBuildkitdImage},
		&container.HostConfig{
			Privileged:    true,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped},
		},
		nil, nil, localBuildkitdContainerName,
	)
	// Conflict: a concurrent werf process created the container first.
	if err != nil && !cerrdefs.IsConflict(err) {
		return fmt.Errorf("create container %q: %w", localBuildkitdContainerName, err)
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
