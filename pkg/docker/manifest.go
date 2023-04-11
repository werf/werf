package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/manifest"
	"golang.org/x/net/context"
)

func doCliManifest(c command.Cli, args ...string) error {
	return prepareCliCmd(manifest.NewManifestCommand(c), args...).Execute()
}

func CliManifest(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliManifest(c, args...)
	})
}
