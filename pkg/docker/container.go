package docker

import (
	"fmt"

	"github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func ContainerInspect(ref string) (types.ContainerJSON, error) {
	ctx := context.Background()
	return apiClient.ContainerInspect(ctx, ref)
}

func ContainerCreate(args []string) error {
	cmd := container.NewCreateCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ContainerRun(args []string) error {
	cmd := container.NewRunCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func ContainerCommit(ref string, commitOptions types.ContainerCommitOptions) (string, error) {
	ctx := context.Background()
	response, err := apiClient.ContainerCommit(ctx, ref, commitOptions)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerRemove(ref string) error {
	ctx := context.Background()
	err := apiClient.ContainerRemove(ctx, ref, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}
