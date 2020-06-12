package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"

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
	liveOutputCli *command.DockerCli
	apiClient     *client.Client

	liveCliOutputEnabled bool
	isDebug              bool
	isVerbose            bool
)

func Init(dockerConfigDir string, verbose, debug bool) error {
	if dockerConfigDir != "" {
		cliconfig.SetDir(dockerConfigDir)
	}

	if err := os.Setenv("DOCKER_CONFIG", dockerConfigDir); err != nil {
		return fmt.Errorf("cannot set DOCKER_CONFIG to %s: %s", dockerConfigDir, err)
	}

	if err := setDockerClient(); err != nil {
		return err
	}

	if err := setDockerApiClient(); err != nil {
		return err
	}

	logrus.StandardLogger().SetOutput(logboek.GetOutStream())

	isDebug = debug
	isVerbose = verbose
	liveCliOutputEnabled = verbose || debug

	return nil
}

func ServerVersion() (*types.Version, error) {
	ctx := context.Background()
	version, err := apiClient.ServerVersion(ctx)
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

func setDockerClient() error {
	if c, err := newDockerCli([]command.DockerCliOption{
		command.WithOutputStream(logboek.GetOutStream()),
		command.WithErrorStream(logboek.GetErrStream()),
		command.WithContentTrust(false),
	}); err != nil {
		return fmt.Errorf("unable to create live output docker cli: %s", err)
	} else {
		liveOutputCli = c
	}

	return nil
}

func setDockerApiClient() error {
	ctx := context.Background()
	serverVersion, err := liveOutputCli.Client().ServerVersion(ctx)
	if err != nil {
		return err
	}

	apiClient, err = client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return err
	}

	return nil
}

func Debug() bool {
	return os.Getenv("WERF_DEBUG_DOCKER") == "1"
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

func callCliWithAutoOutput(commandCaller func(c *command.DockerCli) error) error {
	if liveCliOutputEnabled {
		return commandCaller(liveOutputCli)
	} else {
		output, err := callCliWithRecordedOutput(func(c *command.DockerCli) error {
			return commandCaller(c)
		})
		if err != nil {
			logboek.LogErrorF("%s", output)
		}
		return err
	}
}
