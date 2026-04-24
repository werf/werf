package docker

import (
	"bytes"
	"context"

	"github.com/docker/cli/cli/command"

	"github.com/werf/logboek"
)

func Login(ctx context.Context, username, password, repo string) error {
	var outb, errb bytes.Buffer

	return cliWithCustomOptions(
		ctx,
		[]command.CLIOption{
			command.WithInputStream(nil),
			command.WithOutputStream(&outb),
			command.WithErrorStream(&errb),
		},
		func(c command.Cli) error {
			loginCmd, err := lookupCliCommand(c, "login")
			if err != nil {
				return err
			}
			args := []string{"--username", username, "--password", password, repo}
			cmd := prepareCliCmd(ctx, loginCmd, args...)
			err = cmd.Execute()

			logboek.Context(ctx).Debug().LogF("Docker login stdout:\n%s\nDocker login stderr:\n%s\n", outb.String(), errb.String())

			return err
		},
	)
}
