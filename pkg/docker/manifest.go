package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/manifest"
	"golang.org/x/net/context"
)

func doCliManifest(ctx context.Context, c command.Cli, args ...string) error {
	return prepareCliCmd(ctx, manifest.NewManifestCommand(c), args...).Execute()
}

func CliManifest(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliManifest(ctx, c, args...)
	})
}
