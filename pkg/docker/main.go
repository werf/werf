package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"golang.org/x/net/context"

	"github.com/werf/logboek"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	liveCliOutputEnabled bool
	isDebug              bool
)

func Init(dockerConfigDir string, verbose, debug bool) error {
	if dockerConfigDir != "" {
		cliconfig.SetDir(dockerConfigDir)
	}

	if err := os.Setenv("DOCKER_CONFIG", dockerConfigDir); err != nil {
		return fmt.Errorf("cannot set DOCKER_CONFIG to %s: %s", dockerConfigDir, err)
	}

	isDebug = debug
	liveCliOutputEnabled = verbose || debug

	return nil
}

func ServerVersion() (*types.Version, error) {
	cli, err := newDockerCli([]command.DockerCliOption{
		command.WithContentTrust(false),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create docker cli: %s", err)
	}

	ctx := context.Background()
	version, err := cli.Client().ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func newDockerCli(opts []command.DockerCliOption) (*command.DockerCli, error) {
	newCli, err := command.NewDockerCli(opts...)
	if err != nil {
		return nil, err
	}

	clientOpts := flags.NewClientOptions()
	if isDebug {
		clientOpts.Common.LogLevel = "debug"
	} else {
		clientOpts.Common.LogLevel = "fatal"
	}

	if err := newCli.Initialize(clientOpts); err != nil {
		return nil, err
	}
	return newCli, nil
}

func cli(ctx context.Context) (*command.DockerCli, error) {
	c, err := newDockerCli([]command.DockerCliOption{
		command.WithOutputStream(logboek.Context(ctx).ProxyOutStream()),
		command.WithErrorStream(logboek.Context(ctx).ProxyErrStream()),
		command.WithContentTrust(false),
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create live output docker cli: %s", err)
	}

	return c, nil
}

func apiCli() (*client.Client, error) {
	serverVersion, err := ServerVersion()
	if err != nil {
		return nil, err
	}

	apiClient, err := client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}

func callCliWithRecordedOutput(commandCaller func(c *command.DockerCli) error) (string, error) {
	var output bytes.Buffer

	if c, err := getRecordingOutputCli(&output, &output); err != nil {
		return "", fmt.Errorf("unable to create docker cli: %s", err)
	} else {
		if err := commandCaller(c); err != nil {
			return output.String(), err
		}
		return output.String(), err
	}
}

func getRecordingOutputCli(stdoutWriter, stderrWriter io.Writer) (*command.DockerCli, error) {
	return newDockerCli([]command.DockerCliOption{
		command.WithOutputStream(stdoutWriter),
		command.WithErrorStream(stderrWriter),
		command.WithContentTrust(false),
	})
}

func prepareCliCmd(cmd *cobra.Command, args ...string) *cobra.Command {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	return cmd
}

func callCliWithAutoOutput(ctx context.Context, commandCaller func(c *command.DockerCli) error) error {
	if liveCliOutputEnabled {
		c, err := cli(ctx)
		if err != nil {
			return err
		}

		return commandCaller(c)
	} else {
		output, err := callCliWithRecordedOutput(func(c *command.DockerCli) error {
			return commandCaller(c)
		})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("%s", output)
		}
		return err
	}
}
