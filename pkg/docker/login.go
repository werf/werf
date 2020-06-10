package docker

import (
	"bytes"
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/registry"
	"github.com/docker/cli/cli/flags"
	"github.com/werf/logboek"
)

func Login(username, password, repo string) error {
	var outb, errb bytes.Buffer

	cliOpts := []command.DockerCliOption{
		command.WithInputStream(nil),
		command.WithOutputStream(&outb),
		command.WithErrorStream(&errb),
		command.WithContentTrust(false),
	}

	loginCli, err := command.NewDockerCli(cliOpts...)
	if err != nil {
		return err
	}

	opts := flags.NewClientOptions()
	if err := loginCli.Initialize(opts); err != nil {
		return err
	}

	cmd := registry.NewLoginCommand(loginCli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--username", username, "--password", password, repo})

	err = cmd.Execute()
	if Debug() {
		fmt.Fprintf(logboek.GetOutStream(), "Docker login stdout:\n%s\nDocker login stderr:\n%s\n", outb.String(), errb.String())
	}

	if err != nil {
		return err
	}

	return nil
}
