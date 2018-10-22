package docker

import (
	"github.com/docker/cli/cli/command/registry"
)

func CliLogin(args ...string) error {
	cmd := registry.NewLoginCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}
