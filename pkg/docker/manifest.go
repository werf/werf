package docker

import (
	"github.com/docker/cli/cli/command"
	"golang.org/x/net/context"
)

func doCliManifest(ctx context.Context, c command.Cli, args ...string) error {
	cmd, err := lookupCliCommand(c, "manifest")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
}

func CliManifest(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliManifest(ctx, c, args...)
	})
}
