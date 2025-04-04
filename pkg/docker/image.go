package docker

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
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

type (
	ImagesPruneOptions prune.Options
	ImagesPruneReport  prune.Report
)

// ImagesPrune containers using opts.Filters.
// List of accepted filters is there https://github.com/moby/moby/blob/25.0/daemon/containerd/image_prune.go#L22
func ImagesPrune(ctx context.Context, opts ImagesPruneOptions) (ImagesPruneReport, error) {
	report, err := apiCli(ctx).ImagesPrune(ctx, mapBackendFiltersToImagesPruneFilters(opts.Filters))
	if err != nil {
		return ImagesPruneReport{}, err
	}
	itemsDeleted := lo.Map(report.ImagesDeleted, func(item dockerImage.DeleteResponse, _ int) string {
		return item.Deleted
	})
	return ImagesPruneReport{
		ItemsDeleted:   itemsDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
	}, err
}

func mapBackendFiltersToImagesPruneFilters(list filter.FilterList) filters.Args {
	args := lo.Map(list, func(filter filter.Filter, _ int) filters.KeyValuePair {
		return filters.Arg(filter.First, filter.Second)
	})
	return filters.NewArgs(args...)
}

func doCliPull(c command.Cli, args ...string) error {
	return prepareCliCmd(image.NewPullCommand(c), args...).Execute()
}

func CliPull(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliPull(c, args...)
	})
}

const cliPullMaxAttempts uint8 = 5

func doCliPullWithRetries(ctx context.Context, c command.Cli, args ...string) error {
	var attempt uint8
	op := func() (bool, error) {
		return false, doCliPull(c, args...)
	}
	notify := func(err error, duration time.Duration) {
		logboek.Context(ctx).Warn().LogF("Retrying docker pull in %0.2f seconds (%d/%d) ...\n", duration.Seconds(), attempt, cliPullMaxAttempts)
	}
	return doCliOperationWithRetries(ctx, op, &attempt, cliPullMaxAttempts, notify)
}

func doCliOperationWithRetries(ctx context.Context, op backoff.Operation[bool], opAttempt *uint8, opMaxAttempts uint8, notify backoff.Notify) error {
	isTemporaryErrorMessage := func(errMsg string) bool {
		return slices.ContainsFunc([]string{
			"Client.Timeout exceeded while awaiting headers",
			"TLS handshake timeout",
			"i/o timeout",
			"Only schema version 2 is supported",
			"429 Too Many Requests",
			"504 Gateway Time-out",
			"504 Gateway Timeout",
			"Internal Server Error",
			"authentication required",
		}, func(msgPart string) bool {
			return strings.Contains(errMsg, msgPart)
		})
	}

	opWrapper := func() (bool, error) {
		*opAttempt++
		_, err := op()
		if err != nil {
			if isTemporaryErrorMessage(err.Error()) {
				return false, err
			}
			// Do not retry on other errors.
			return false, backoff.Permanent(err)
		}
		return false, nil
	}

	eb := backoff.NewExponentialBackOff()
	eb.MaxInterval = 30 * time.Second

	_, err := backoff.Retry(ctx, opWrapper,
		backoff.WithBackOff(eb),
		backoff.WithMaxTries(uint(opMaxAttempts)),
		backoff.WithNotify(notify))
	if err != nil {
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

const cliPushMaxAttempts uint8 = 10

func doCliPushWithRetries(ctx context.Context, c command.Cli, args ...string) error {
	var attempt uint8
	op := func() (bool, error) {
		err := doCliPush(c, args...)
		return false, err
	}
	notify := func(err error, duration time.Duration) {
		logboek.Context(ctx).Warn().LogF("Retrying docker push in %0.2f seconds (%d/%d) ...\n", duration.Seconds(), attempt, cliPushMaxAttempts)
	}
	return doCliOperationWithRetries(ctx, op, &attempt, cliPushMaxAttempts, notify)
}

func CliPushWithRetries(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliPushWithRetries(ctx, c, args...)
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
		finalArgs = append([]string{"build", "--load"}, args...)
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

		// TODO: --provenance=false is a workaround for index manifests that we cannot handle properly with current code base (fix in v3).
		args = append([]string{"--provenance=false"}, args...)
	} else {
		var err error
		args, err = checkForUnsupportedOptions(ctx, args...)
		if err != nil {
			return err
		}
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

func checkForUnsupportedOptions(ctx context.Context, args ...string) ([]string, error) {
	borderIndex := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.Contains(arg, "--secret") {
			return nil, fmt.Errorf("secrets are only available with Docker BuildKit")
		}
		// since we use our ssh agent as default we need to pop ssh options
		// to be able to run build with legacy backend
		if strings.Contains(arg, "--ssh") {
			logboek.Context(ctx).Info().LogF("--ssh is not supported by legacy Docker builder so it will be skipped")
			if i+1 < len(args) {
				i++
			}
			continue
		}
		args[borderIndex] = args[i]
		borderIndex++
	}
	return args[:borderIndex], nil
}
