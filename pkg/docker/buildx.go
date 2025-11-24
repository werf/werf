package docker

import (
	"github.com/docker/buildx/commands"
	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

func NewBuildxCommand(dockerCli command.Cli) *cobra.Command {
	// plugin.PersistentPreRunE = func(*cobra.Command, []string) error { return nil }
	cmd := commands.NewRootCmd("", false, dockerCli)

	// ---------------
	// IMPORTANT: buildx uses this function to hardcode ctx (appcontext) which handles system signals.
	// https://github.com/werf/3p-docker-buildx/blob/v0.13.0-rc2/commands/root.go#L34
	// It is a workaround to be able to pass ctx for cancellation of command execution.
	// ---------------
	cmd.PersistentPreRunE = nil

	return cmd
}
