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
	defaultPlatform      string
	runtimePlatform      string
	useBuildx            bool

	DockerConfigDir string
)

const (
	ctxDockerCliKey = "docker_cli"
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

	isDebug = os.Getenv("WERF_DEBUG_DOCKER") == "1"
	liveCliOutputEnabled = opts.Verbose || opts.Debug

	defaultCLI, err = newDockerCli(defaultCliOptions(ctx))
	if err != nil {
		return err
	}

	spec := platforms.DefaultSpec()
	spec.OS = defaultCLI.ServerInfo().OSType
	runtimePlatform = platforms.Format(spec)
	claimPlatforms := opts.ClaimPlatforms

	if opts.DefaultPlatform != "" {
		defaultPlatform = opts.DefaultPlatform
		os.Setenv("DOCKER_DEFAULT_PLATFORM", opts.DefaultPlatform)
		claimPlatforms = append(claimPlatforms, opts.DefaultPlatform)
	} else {
		defaultPlatform = runtimePlatform
	}

	for _, claimPlatform := range claimPlatforms {
		if claimPlatform != runtimePlatform {
			useBuildx = true
			break
		}
	}

	if v := os.Getenv("DOCKER_BUILDKIT"); v == "1" || v == "true" {
		if err := os.Setenv("DOCKER_BUILDKIT", "0"); err != nil {
			return fmt.Errorf("unable to set env var: %w", err)
		}
		useBuildx = true
	}

	return nil
}

func ClaimTargetPlatforms(claimPlatforms []string) {
	if defaultPlatform != "" {
		claimPlatforms = append(claimPlatforms, defaultPlatform)
	}
	for _, claimPlatform := range claimPlatforms {
		if claimPlatform != runtimePlatform {
			useBuildx = true
			break
		}
	}
}

func GetDefaultPlatform() string {
	return defaultPlatform
}

func GetRuntimePlatform() string {
	return runtimePlatform
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
