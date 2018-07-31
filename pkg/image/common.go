package image

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	commandContainer "github.com/docker/cli/cli/command/container"
	commandImage "github.com/docker/cli/cli/command/image"
)

func Pull(name string, cli *command.DockerCli) error {
	cmd := commandImage.NewPullCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Push(name string, cli *command.DockerCli) error {
	cmd := commandImage.NewPushCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Tag(ref, tag string, cli *command.DockerCli) error {
	cmd := commandImage.NewTagCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{ref, tag})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Untag(tag string, cli *command.DockerCli) error {
	cmd := commandImage.NewRemoveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{tag})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Save(images []string, filePath string, cli *command.DockerCli) error {
	var args []string
	args = append(args, []string{"-o", filePath}...)
	args = append(args, images...)

	cmd := commandImage.NewSaveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Load(filePath string, cli *command.DockerCli) error {
	cmd := commandImage.NewLoadCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"-i", filePath})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ContainerRun(args []string, cli *command.DockerCli) error {
	cmd := commandContainer.NewRunCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}
