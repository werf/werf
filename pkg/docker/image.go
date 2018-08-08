package docker

import (
	"github.com/docker/cli/cli/command/image"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func ImageInspectWithRaw(ref string) (*types.ImageInspect, error) {
	ctx := context.Background()
	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}

func ImagePull(name string) error {
	cmd := image.NewPullCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ImagePush(name string) error {
	cmd := image.NewPushCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ImageTag(ref, tag string) error {
	cmd := image.NewTagCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{ref, tag})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ImageUntag(tag string) error {
	cmd := image.NewRemoveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{tag})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ImagesSave(images []string, filePath string) error {
	var args []string
	args = append(args, []string{"-o", filePath}...)
	args = append(args, images...)

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

func ImagesLoad(filePath string) error {
	cmd := image.NewLoadCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"-i", filePath})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}
