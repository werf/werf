package docker

import (
	"fmt"
	"io"
	"strings"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/cli/cli/command"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/versions"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/net/context"

	"github.com/werf/werf/v3/pkg/opstats"
)

func Containers(ctx context.Context, options client.ContainerListOptions) ([]dockercontainer.Summary, error) {
	defer opstats.Observe(ctx, "docker: container list")()

	result, err := apiCli(ctx).ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

func ContainerExist(ctx context.Context, ref string) (bool, error) {
	if _, err := ContainerInspect(ctx, ref); err != nil {
		if cerrdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ContainerAttach(ctx context.Context, ref string, options client.ContainerAttachOptions) (client.HijackedResponse, error) {
	result, err := apiCli(ctx).ContainerAttach(ctx, ref, options)
	if err != nil {
		return client.HijackedResponse{}, err
	}

	return result.HijackedResponse, nil
}

func ContainerInspect(ctx context.Context, ref string) (dockercontainer.InspectResponse, error) {
	defer opstats.Observe(ctx, "docker: container inspect")()

	result, err := apiCli(ctx).ContainerInspect(ctx, ref, client.ContainerInspectOptions{})
	if err != nil {
		return dockercontainer.InspectResponse{}, err
	}

	return result.Container, nil
}

func ContainerCreate(ctx context.Context, config *dockercontainer.Config, platform *ocispec.Platform, name string) (string, error) {
	defer opstats.Observe(ctx, "docker: container create")()

	if err := CheckConnection(ctx, CheckConnectionOptions{}); err != nil {
		return "", err
	}
	api := apiCli(ctx)
	version := api.ClientVersion()
	if platform != nil && versions.LessThan(version, "1.41") {
		return "", fmt.Errorf("%q requires API version 1.41, but the Docker daemon API version is %s", "specify container image platform", version)
	}
	if config != nil && config.Healthcheck != nil && config.Healthcheck.StartInterval != 0 && versions.LessThan(version, "1.44") {
		return "", fmt.Errorf("%q requires API version 1.44, but the Docker daemon API version is %s", "specify health-check start interval", version)
	}

	response, err := api.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:           config,
		NetworkingConfig: &network.NetworkingConfig{},
		Platform:         platform,
		Name:             name,
	})
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerCommit(ctx context.Context, ref string, commitOptions client.ContainerCommitOptions) (string, error) {
	defer opstats.Observe(ctx, "docker: container commit")()

	response, err := apiCli(ctx).ContainerCommit(ctx, ref, commitOptions)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerRemove(ctx context.Context, ref string, options client.ContainerRemoveOptions) error {
	defer opstats.Observe(ctx, "docker: container remove")()

	_, err := apiCli(ctx).ContainerRemove(ctx, ref, options)
	return err
}

func doCliCreate(ctx context.Context, c command.Cli, args ...string) error {
	defer opstats.Observe(ctx, "docker: container create")()

	cmd, err := lookupCliCommand(c, "create")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
}

func CliCreate(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliCreate(ctx, c, args...)
	})
}

func doCliRun(ctx context.Context, c command.Cli, args ...string) error {
	defer opstats.Observe(ctx, "docker: container run")()

	cmd, err := lookupCliCommand(c, "run")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
}

func CliRun(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliRun(ctx, c, args...)
	})
}

func CliRun_ProvidedOutput(ctx context.Context, stdoutWriter, stderrWriter io.Writer, args ...string) error {
	return callCliWithProvidedOutput(ctx, stdoutWriter, stderrWriter, func(c command.Cli) error {
		return doCliRun(ctx, c, args...)
	})
}

func CliRun_LiveOutput(ctx context.Context, args ...string) error {
	return doCliRun(ctx, cli(ctx), args...)
}

func CliRunWithInput_LiveOutput(ctx context.Context, input string, args ...string) error {
	return cliWithCustomOptions(ctx, []command.CLIOption{command.WithInputStream(io.NopCloser(strings.NewReader(input)))}, func(c command.Cli) error {
		return doCliRun(ctx, c, args...)
	})
}

func CliRun_RecordedOutput(ctx context.Context, args ...string) (string, error) {
	return callCliWithRecordedOutput(ctx, func(c command.Cli) error {
		return doCliRun(ctx, c, args...)
	})
}

func doCliRm(ctx context.Context, c command.Cli, args ...string) error {
	defer opstats.Observe(ctx, "docker: container remove")()

	cmd, err := lookupCliCommand(c, "rm")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
}

func CliRm(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliRm(ctx, c, args...)
	})
}

func CliRm_RecordedOutput(ctx context.Context, args ...string) (string, error) {
	return callCliWithRecordedOutput(ctx, func(c command.Cli) error {
		return doCliRm(ctx, c, args...)
	})
}
