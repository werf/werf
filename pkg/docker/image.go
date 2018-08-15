package docker

import (
	"github.com/docker/cli/cli/command/image"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func ImageInspect(ref string) (*types.ImageInspect, error) {
	ctx := context.Background()
	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}

func CliPull(args ...string) error {
	cmd := image.NewPullCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliPush(args ...string) error {
	cmd := image.NewPushCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliTag(args ...string) error {
	cmd := image.NewTagCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliRmi(args ...string) error {
	cmd := image.NewRemoveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliSave(args ...string) error {
	cmd := image.NewSaveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliLoad(args ...string) error {
	cmd := image.NewLoadCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}
