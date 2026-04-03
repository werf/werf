package docker

import (
	"fmt"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/commands"
	"github.com/spf13/cobra"
)

var requiredCliCommands = []string{
	"create",
	"login",
	"manifest",
	"pull",
	"push",
	"rm",
	"rmi",
	"run",
	"tag",
}

func lookupCliCommand(c command.Cli, name string) (*cobra.Command, error) {
	root := &cobra.Command{Use: "docker"}
	commands.AddCommands(root, c)
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			root.RemoveCommand(cmd)
			return cmd, nil
		}
	}
	return nil, fmt.Errorf("docker CLI command %q not found", name)
}

func validateRequiredCliCommands(c command.Cli) error {
	root := &cobra.Command{Use: "docker"}
	commands.AddCommands(root, c)

	available := make(map[string]struct{}, len(root.Commands()))
	for _, cmd := range root.Commands() {
		available[cmd.Name()] = struct{}{}
	}

	var missing []string
	for _, name := range requiredCliCommands {
		if _, ok := available[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required docker CLI commands not found: %s", strings.Join(missing, ", "))
	}
	return nil
}
