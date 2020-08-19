package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func Containers(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	apiClient, err := apiCli()
	if err != nil {
		return nil, err
	}

	return apiClient.ContainerList(ctx, options)
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

func ContainerInspect(ctx context.Context, ref string) (types.ContainerJSON, error) {
	apiClient, err := apiCli()
	if err != nil {
		return types.ContainerJSON{}, err
	}

	return apiClient.ContainerInspect(ctx, ref)
}

func ContainerCommit(ctx context.Context, ref string, commitOptions types.ContainerCommitOptions) (string, error) {
	apiClient, err := apiCli()
	if err != nil {
		return "", err
	}

	response, err := apiClient.ContainerCommit(ctx, ref, commitOptions)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerRemove(ctx context.Context, ref string, options types.ContainerRemoveOptions) error {
	apiClient, err := apiCli()
	if err != nil {
		return err
	}

	return apiClient.ContainerRemove(ctx, ref, options)
}

func doCliCreate(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(container.NewCreateCommand(c), args...).Execute()
}

func CliCreate(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliCreate(c, args...)
	})
}

func CliCreate_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliCreate(cli, args...)
}

func CliCreate_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliCreate(c, args...)
	})
}

func doCliRun(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(container.NewRunCommand(c), args...).Execute()
}

func CliRun(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliRun(c, args...)
	})
}

func CliRun_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliRun(cli, args...)
}

func CliRun_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliRun(c, args...)
	})
}

func doCliRm(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(container.NewRmCommand(c), args...).Execute()
}

func CliRm(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliRm(c, args...)
	})
}

func CliRm_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliRm(cli, args...)
}

func CliRm_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliRm(c, args...)
	})
}
