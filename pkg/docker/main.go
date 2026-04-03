package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/platforms"
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
	defaultCLI           command.Cli
	defaultAPIClient     client.APIClient
	defaultPlatform      string
	runtimePlatform      string

	DockerConfigDir string
)

const (
	ctxDockerCliKey = "docker_cli"
	ctxAPIClientKey = "docker_api_client"
)

func IsEnabled() bool {
	return defaultCLI != nil
}

type InitOptions struct {
	DockerConfigDir string
	DefaultPlatform string
	ClaimPlatforms  []string
	Verbose         bool
	Debug           bool
}

func Init(ctx context.Context, opts InitOptions) error {
	err := InitDockerConfig(opts)
	if err != nil {
		return err
	}

	isDebug = os.Getenv("WERF_DEBUG_DOCKER") == "1"
	liveCliOutputEnabled = opts.Verbose || opts.Debug

	defaultCLI, err = newDockerCli(defaultCliOptions(ctx))
	if err != nil {
		return err
	}

	defaultAPIClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	spec := platforms.DefaultSpec()
	spec.OS = defaultCLI.ServerInfo().OSType
	runtimePlatform = platforms.Format(spec)

	if opts.DefaultPlatform != "" {
		defaultPlatform = opts.DefaultPlatform
		os.Setenv("DOCKER_DEFAULT_PLATFORM", opts.DefaultPlatform)
	} else {
		defaultPlatform = runtimePlatform
	}

	return nil
}

func InitDockerConfig(opts InitOptions) error {
	if opts.DockerConfigDir != "" {
		DockerConfigDir = opts.DockerConfigDir
		cliconfig.SetDir(opts.DockerConfigDir)
	} else {
		DockerConfigDir = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	err := os.Setenv("DOCKER_CONFIG", DockerConfigDir)
	if err != nil {
		return fmt.Errorf("cannot set DOCKER_CONFIG to %s: %w", DockerConfigDir, err)
	}

	return nil
}

func ClaimTargetPlatforms(claimPlatforms []string) {
}

func GetDefaultPlatform() string {
	return defaultPlatform
}

func GetRuntimePlatform() string {
	return runtimePlatform
}

func ServerVersion(ctx context.Context) (*types.Version, error) {
	version, err := apiCli(ctx).ServerVersion(ctx)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func newDockerCli(opts []command.CLIOption) (command.Cli, error) {
	newCli, err := command.NewDockerCli(opts...)
	if err != nil {
		return nil, err
	}

	clientOpts := flags.NewClientOptions()

	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.Dir()
	}

	clientOpts.TLS = os.Getenv("DOCKER_TLS") != ""
	clientOpts.TLSVerify = os.Getenv("DOCKER_TLS_VERIFY") != ""

	if clientOpts.TLSVerify {
		clientOpts.TLSOptions = &tlsconfig.Options{
			CAFile:   filepath.Join(dockerCertPath, flags.DefaultCaFile),
			CertFile: filepath.Join(dockerCertPath, flags.DefaultCertFile),
			KeyFile:  filepath.Join(dockerCertPath, flags.DefaultKeyFile),
		}
	}

	if isDebug {
		clientOpts.LogLevel = "debug"
	} else {
		clientOpts.LogLevel = "fatal"
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
		return defaultCLI
	default:
		panic("context is not bound with docker cli")
	}
}

func apiCli(ctx context.Context) client.APIClient {
	apiClientInterf := ctx.Value(ctxAPIClientKey)
	switch {
	case apiClientInterf != nil:
		return apiClientInterf.(client.APIClient)
	case ctx == context.Background():
		return defaultAPIClient
	default:
		return defaultAPIClient
	}
}

func defaultCliOptions(ctx context.Context) []command.CLIOption {
	return []command.CLIOption{
		command.WithInputStream(os.Stdin),
		command.WithOutputStream(logboek.Context(ctx).OutStream()),
		command.WithErrorStream(logboek.Context(ctx).ErrStream()),
	}
}

func applyCliOptions(c command.Cli, options []command.CLIOption) error {
	dockerCli, ok := c.(*command.DockerCli)
	if !ok {
		return fmt.Errorf("expected *command.DockerCli, got %T", c)
	}
	for _, opt := range options {
		if err := opt(dockerCli); err != nil {
			return err
		}
	}
	return nil
}

func cliWithCustomOptions(ctx context.Context, options []command.CLIOption, f func(cli command.Cli) error) error {
	if err := applyCliOptions(cli(ctx), options); err != nil {
		return err
	}

	err := f(cli(ctx))

	if applyErr := applyCliOptions(cli(ctx), defaultCliOptions(ctx)); applyErr != nil {
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

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("unable to create docker api client: %w", err)
	}

	newCtx := context.WithValue(ctx, ctxDockerCliKey, c)
	newCtx = context.WithValue(newCtx, ctxAPIClientKey, apiClient)
	return newCtx, nil
}

func IsContext(ctx context.Context) bool {
	return ctx.Value(ctxDockerCliKey) != nil
}

func callCliWithProvidedOutput(ctx context.Context, stdoutWriter, stderrWriter io.Writer, commandCaller func(c command.Cli) error) error {
	var errOutput bytes.Buffer

	if err := cliWithCustomOptions(
		ctx,
		[]command.CLIOption{
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
		[]command.CLIOption{
			command.WithOutputStream(&output),
			command.WithErrorStream(&output),
		},
		commandCaller,
	); err != nil {
		return "", err
	}

	return output.String(), nil
}

func prepareCliCmd(ctx context.Context, cmd *cobra.Command, args ...string) *cobra.Command {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
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
