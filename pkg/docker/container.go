package docker

import (
	"io"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/werf/werf/v2/pkg/opstats"
)

func Containers(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	defer opstats.Observe(ctx, "docker: container list")()

	return apiCli(ctx).ContainerList(ctx, options)
}

func ContainerExist(ctx context.Context, ref string) (bool, error) {
	if _, err := ContainerInspect(ctx, ref); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ContainerAttach(ctx context.Context, ref string, options types.ContainerAttachOptions) (types.HijackedResponse, error) {
	// Measures the attach call itself: the observation ends when the hijacked
	// connection is handed to the caller, not when the streams are done with.
	defer opstats.Observe(ctx, "docker: container attach")()

	return apiCli(ctx).ContainerAttach(ctx, ref, options)
}

func ContainerInspect(ctx context.Context, ref string) (types.ContainerJSON, error) {
	defer opstats.Observe(ctx, "docker: container inspect")()

	return apiCli(ctx).ContainerInspect(ctx, ref)
}

func ContainerCommit(ctx context.Context, ref string, commitOptions types.ContainerCommitOptions) (string, error) {
	defer opstats.Observe(ctx, "docker: container commit")()

	response, err := apiCli(ctx).ContainerCommit(ctx, ref, commitOptions)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerRemove(ctx context.Context, ref string, options types.ContainerRemoveOptions) error {
	defer opstats.Observe(ctx, "docker: container remove")()

	return apiCli(ctx).ContainerRemove(ctx, ref, options)
}

func doCliCreate(ctx context.Context, c command.Cli, args ...string) error {
	defer opstats.Observe(ctx, "docker: container create")()

	return prepareCliCmd(ctx, container.NewCreateCommand(c), args...).Execute()
}

func CliCreate(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliCreate(ctx, c, args...)
	})
}

func doCliRun(ctx context.Context, c command.Cli, args ...string) error {
	defer opstats.Observe(ctx, "docker: container run")()

	return prepareCliCmd(ctx, container.NewRunCommand(c), args...).Execute()
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

	return prepareCliCmd(ctx, container.NewRmCommand(c), args...).Execute()
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
