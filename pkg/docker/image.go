package docker

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/werf/logboek"
	parallelConstant "github.com/werf/werf/pkg/util/parallel/constant"
)

type CreateImageOptions struct {
	Labels []string
}

func CreateImage(ctx context.Context, ref string, opts CreateImageOptions) error {
	var importOpts types.ImageImportOptions
	if len(opts.Labels) > 0 {
		changeOption := "LABEL"
		for _, label := range opts.Labels {
			changeOption += fmt.Sprintf(" %s", label)
		}
		importOpts.Changes = append(importOpts.Changes, changeOption)
	}
	_, err := apiCli(ctx).ImageImport(ctx, types.ImageImportSource{SourceName: "-"}, ref, importOpts)
	return err
}

func Images(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	images, err := apiCli(ctx).ImageList(ctx, options)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func ImageExist(ctx context.Context, ref string) (bool, error) {
	if _, err := ImageInspect(ctx, ref); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ImageInspect(ctx context.Context, ref string) (*types.ImageInspect, error) {
	inspect, _, err := apiCli(ctx).ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}

func doCliPull(c command.Cli, args ...string) error {
	return prepareCliCmd(image.NewPullCommand(c), args...).Execute()
}

func CliPull(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliPull(c, args...)
	})
}

const cliPullMaxAttempts = 5

func doCliPullWithRetries(ctx context.Context, c command.Cli, args ...string) error {
	var attempt int

tryPull:
	if err := doCliPull(c, args...); err != nil {
		if attempt < cliPullMaxAttempts {
			specificErrors := []string{
				"Client.Timeout exceeded while awaiting headers",
				"TLS handshake timeout",
				"i/o timeout",
				"504 Gateway Time-out",
				"504 Gateway Timeout",
				"Internal Server Error",
			}

			for _, specificError := range specificErrors {
				if strings.Contains(err.Error(), specificError) {
					attempt++
					seconds := rand.Intn(30-15) + 15 // from 15 to 30 seconds

					logboek.Context(ctx).Warn().LogF("Retrying docker pull in %d seconds (%d/%d) ...\n", seconds, attempt, cliPullMaxAttempts)
					time.Sleep(time.Duration(seconds) * time.Second)
					goto tryPull
				}
			}
		}

		return err
	}

	return nil
}

func CliPullWithRetries(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliPullWithRetries(ctx, c, args...)
	})
}

func doCliPush(c command.Cli, args ...string) error {
	return prepareCliCmd(image.NewPushCommand(c), args...).Execute()
}

const cliPushMaxAttempts = 10

func doCliPushWithRetries(c command.Cli, args ...string) error {
	var attempt int

tryPush:
	if err := doCliPush(c, args...); err != nil {
		if attempt < cliPushMaxAttempts {
			specificErrors := []string{
				"Client.Timeout exceeded while awaiting headers",
				"TLS handshake timeout",
				"i/o timeout",
				"Only schema version 2 is supported",
				"504 Gateway Time-out",
				"504 Gateway Timeout",
				"Internal Server Error",
			}

			for _, specificError := range specificErrors {
				if strings.Contains(err.Error(), specificError) {
					attempt++
					seconds := rand.Intn(30-15) + 15 // from 15 to 30 seconds

					msg := fmt.Sprintf("Retrying docker push in %d seconds (%d/%d) ...\n", seconds, attempt, cliPushMaxAttempts)
					_, _ = c.Err().Write([]byte(msg))

					time.Sleep(time.Duration(seconds) * time.Second)
					goto tryPush
				}
			}
		}

		return err
	}

	return nil
}

func CliPushWithRetries(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliPushWithRetries(c, args...)
	})
}

func doCliTag(c command.Cli, args ...string) error {
	return prepareCliCmd(image.NewTagCommand(c), args...).Execute()
}

func CliTag(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliTag(c, args...)
	})
}

func doCliRmi(c command.Cli, args ...string) error {
	return prepareCliCmd(image.NewRemoveCommand(c), args...).Execute()
}

func CliRmi(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliRmi(c, args...)
	})
}

func CliRmi_LiveOutput(ctx context.Context, args ...string) error {
	return doCliRmi(cli(ctx), args...)
}

type BuildOptions struct {
	EnableBuildx bool
}

func doCliBuild(c command.Cli, opts BuildOptions, args ...string) error {
	var finalArgs []string
	var cmd *cobra.Command

	if opts.EnableBuildx {
		cmd = NewBuildxCommand(c)
		finalArgs = append([]string{"build"}, args...)
	} else {
		cmd = image.NewBuildCommand(c)
		finalArgs = args
	}

	return prepareCliCmd(cmd, finalArgs...).Execute()
}

func CliBuild_LiveOutputWithCustomIn(ctx context.Context, rc io.ReadCloser, args ...string) error {
	buildOpts := BuildOptions{}

	if useBuildx {
		buildOpts.EnableBuildx = true

		// disable buildkit output in background tasks due to https://github.com/docker/cli/issues/2889
		// there is no true way to get output, because buildkit uses the standard output and error streams instead of defined ones in the cli instance
		if ctx.Value(parallelConstant.CtxBackgroundTaskIDKey) != nil {
			logboek.Context(ctx).Warn().LogLn("WARNING: BuildKit output in background tasks is not supported (--quiet) due to https://github.com/docker/cli/issues/2889")
			args = append(args, "--quiet")
		}
	} else {
		// ensure buildkit not enabled
		if err := os.Setenv("DOCKER_BUILDKIT", "0"); err != nil {
			return err
		}
	}

	return cliWithCustomOptions(ctx, []command.DockerCliOption{
		func(cli *command.DockerCli) error {
			cli.SetIn(streams.NewIn(rc))
			return nil
		},
	}, func(cli command.Cli) error {
		return doCliBuild(cli, buildOpts, args...)
	})
}

func CliBuild_LiveOutput(ctx context.Context, args ...string) error {
	buildOpts := BuildOptions{EnableBuildx: useBuildx}
	return doCliBuild(cli(ctx), buildOpts, args...)
}
