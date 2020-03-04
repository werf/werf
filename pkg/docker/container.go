package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func Containers(options types.ContainerListOptions) ([]types.Container, error) {
	ctx := context.Background()
	return apiClient.ContainerList(ctx, options)
}

func ContainerExist(ref string) (bool, error) {
	if _, err := ContainerInspect(ref); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ContainerInspect(ref string) (types.ContainerJSON, error) {
	ctx := context.Background()
	return apiClient.ContainerInspect(ctx, ref)
}

func ContainerCommit(ref string, commitOptions types.ContainerCommitOptions) (string, error) {
	ctx := context.Background()
	response, err := apiClient.ContainerCommit(ctx, ref, commitOptions)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerRemove(ref string, options types.ContainerRemoveOptions) error {
	ctx := context.Background()
	err := apiClient.ContainerRemove(ctx, ref, options)
	if err != nil {
		return err
	}

	return nil
}

func doCliCreate(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(container.NewCreateCommand(c), args...).Execute()
}

func CliCreate(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliCreate(c, args...)
	})
}

func CliCreate_LiveOutput(args ...string) error {
	return doCliCreate(liveOutputCli, args...)
}

func CliCreate_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliCreate(c, args...)
	})
}

func doCliRun(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(container.NewRunCommand(c), args...).Execute()
}

func CliRun(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliRun(c, args...)
	})
}

func CliRun_LiveOutput(args ...string) error {
	return doCliRun(liveOutputCli, args...)
}

func CliRun_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliRun(c, args...)
	})
}

func doCliRm(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(container.NewRmCommand(c), args...).Execute()
}

func CliRm(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliRm(c, args...)
	})
}

func CliRm_LiveOutput(args ...string) error {
	return doCliRm(liveOutputCli, args...)
}

func CliRm_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliRm(c, args...)
	})
}
