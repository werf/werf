package docker

import (
	"github.com/docker/buildx/commands"
	_ "github.com/docker/buildx/driver/docker"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

func NewBuildxCommand(dockerCli command.Cli) *cobra.Command {
	// plugin.PersistentPreRunE = func(*cobra.Command, []string) error { return nil }
	cmd := commands.NewRootCmd("", false, dockerCli)
	return cmd
}
