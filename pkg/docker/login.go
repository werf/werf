package docker

import (
	"bytes"
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/registry"
	"github.com/docker/cli/cli/flags"
)

func Login(username, password, repo string) error {
	var outb, errb bytes.Buffer

	loginCli := command.NewDockerCli(nil, &outb, &errb, false)
	opts := flags.NewClientOptions()
	if err := loginCli.Initialize(opts); err != nil {
		return err
	}

	cmd := registry.NewLoginCommand(loginCli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--username", username, "--password", password, repo})

	err := cmd.Execute()
	if debug() {
		fmt.Printf("Docker login stdout:\n%s\nDocker login stderr:\n%s\n", outb.String(), errb.String())
	}

	if err != nil {
		return err
	}

	return nil
}
