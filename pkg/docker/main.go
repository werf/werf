package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/werf/logboek"
)

var (
	liveCliOutputEnabled bool
	isDebug              bool
	defaultCLi           command.Cli

	DockerConfigDir string
)

const (
	ctxDockerCliKey = "docker_cli"
)

func IsEnabled() bool {
	return defaultCLi != nil
}

func Init(ctx context.Context, dockerConfigDir string, verbose, debug bool, platform string) error {
	if (platform == "" && runtime.GOARCH != "amd64") || (platform != "" && platform != "linux/amd64") {
		logboek.Context(ctx).Error().LogF("werf currently does not support building of images for any other platform besides linux/amd64.\n")
		logboek.Context(ctx).Error().LogF("Please set --platform option (or WERF_PLATFORM, or DOCKER_DEFAULT_PLATFORM environment variable) to linux/amd64 to enable platform emulation when building images with werf.\n")
		logboek.Context(ctx).Error().LogLn()
		return fmt.Errorf("unsupported platform")
	}
	if platform != "" {
		os.Setenv("DOCKER_DEFAULT_PLATFORM", platform)
		os.Setenv("DOCKER_BUILDKIT", "1")
	}

	DockerConfigDir = dockerConfigDir

	if dockerConfigDir == "" {
		DockerConfigDir = filepath.Join(os.Getenv("HOME"), ".docker")
	} else {
		cliconfig.SetDir(dockerConfigDir)
	}

	err := os.Setenv("DOCKER_CONFIG", dockerConfigDir)
	if err != nil {
		return fmt.Errorf("cannot set DOCKER_CONFIG to %s: %w", dockerConfigDir, err)
	}

	isDebug = os.Getenv("WERF_DEBUG_DOCKER") == "1"
	liveCliOutputEnabled = verbose || debug

	defaultCLi, err = newDockerCli(defaultCliOptions(ctx))
	if err != nil {
		return err
	}

	return nil
}

func ServerVersion(ctx context.Context) (*types.Version, error) {
	version, err := cli(ctx).Client().ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func newDockerCli(opts []command.DockerCliOption) (command.Cli, error) {
	newCli, err := command.NewDockerCli(opts...)
	if err != nil {
		return nil, err
	}

	clientOpts := flags.NewClientOptions()

	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.Dir()
	}

	clientOpts.Common.TLS = os.Getenv("DOCKER_TLS") != ""
	clientOpts.Common.TLSVerify = os.Getenv("DOCKER_TLS_VERIFY") != ""

	if clientOpts.Common.TLSVerify {
		clientOpts.Common.TLSOptions = &tlsconfig.Options{
			CAFile:   filepath.Join(dockerCertPath, flags.DefaultCaFile),
			CertFile: filepath.Join(dockerCertPath, flags.DefaultCertFile),
			KeyFile:  filepath.Join(dockerCertPath, flags.DefaultKeyFile),
		}
	}

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

func cli(ctx context.Context) command.Cli {
	cliInterf := ctx.Value(ctxDockerCliKey)
	switch {
	case cliInterf != nil:
		return cliInterf.(command.Cli)
	case ctx == context.Background():
		return defaultCLi
	default:
		panic("context is not bound with docker cli")
	}
}

func apiCli(ctx context.Context) client.APIClient {
	return cli(ctx).Client()
}

func defaultCliOptions(ctx context.Context) []command.DockerCliOption {
	return []command.DockerCliOption{
		command.WithInputStream(os.Stdin),
		command.WithOutputStream(logboek.Context(ctx).OutStream()),
		command.WithErrorStream(logboek.Context(ctx).ErrStream()),
		command.WithContentTrust(false),
	}
}

func cliWithCustomOptions(ctx context.Context, options []command.DockerCliOption, f func(cli command.Cli) error) error {
	if err := cli(ctx).Apply(options...); err != nil {
		return err
	}

	err := f(cli(ctx))

	if applyErr := cli(ctx).Apply(defaultCliOptions(ctx)...); applyErr != nil {
		if err != nil {
			return err
		} else {
			return applyErr
		}
	}

	return err
}

func NewContext(ctx context.Context) (context.Context, error) {
	c, err := newDockerCli(defaultCliOptions(ctx))
	if err != nil {
		return nil, fmt.Errorf("unable to create docker cli: %w", err)
	}

	newCtx := context.WithValue(ctx, ctxDockerCliKey, c)
	return newCtx, nil
}

func IsContext(ctx context.Context) bool {
	return ctx.Value(ctxDockerCliKey) != nil
}

func SyncContextCliWithLogger(ctx context.Context) error {
	return cli(ctx).Apply(defaultCliOptions(ctx)...)
}

func callCliWithProvidedOutput(ctx context.Context, stdoutWriter, stderrWriter io.Writer, commandCaller func(c command.Cli) error) error {
	var errOutput bytes.Buffer

	if err := cliWithCustomOptions(
		ctx,
		[]command.DockerCliOption{
			command.WithOutputStream(stdoutWriter),
			command.WithErrorStream(io.MultiWriter(stderrWriter, &errOutput)),
		},
		commandCaller,
	); err != nil {
		return fmt.Errorf("docker failed:\n%s\n---\n%w", errOutput.String(), err)
	}

	return nil
}

func callCliWithRecordedOutput(ctx context.Context, commandCaller func(c command.Cli) error) (string, error) {
	var output bytes.Buffer

	if err := cliWithCustomOptions(
		ctx,
		[]command.DockerCliOption{
			command.WithOutputStream(&output),
			command.WithErrorStream(&output),
		},
		commandCaller,
	); err != nil {
		return "", err
	}

	return output.String(), nil
}

func prepareCliCmd(cmd *cobra.Command, args ...string) *cobra.Command {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	return cmd
}

func callCliWithAutoOutput(ctx context.Context, commandCaller func(c command.Cli) error) error {
	if liveCliOutputEnabled {
		return commandCaller(cli(ctx))
	} else {
		output, err := callCliWithRecordedOutput(ctx, func(c command.Cli) error {
			return commandCaller(c)
		})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("%s", output)
		}
		return err
	}
}
