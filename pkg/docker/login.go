package docker

import (
	"bytes"
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/registry"

	"github.com/werf/logboek"
)

func Login(ctx context.Context, username, password, repo string) error {
	var outb, errb bytes.Buffer

	return cliWithCustomOptions(
		ctx,
		[]command.DockerCliOption{
			command.WithInputStream(nil),
			command.WithOutputStream(&outb),
			command.WithErrorStream(&errb),
		},
		func(cli command.Cli) error {
			cmd := registry.NewLoginCommand(cli)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			cmd.SetArgs([]string{"--username", username, "--password", password, repo})

			err := cmd.Execute()
			logboek.Context(ctx).Debug().LogF("Docker login stdout:\n%s\nDocker login stderr:\n%s\n", outb.String(), errb.String())

			return err
		},
	)
}
