package docker

import (
	"bytes"
	"fmt"
	"io"
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

func doCliPull(ctx context.Context, c command.Cli, args ...string) error {
	return prepareCliCmd(ctx, image.NewPullCommand(c), args...).Execute()
}

func CliPull(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliPull(ctx, c, args...)
	})
}

const cliPullMaxAttempts uint8 = 5

func doCliPullWithRetries(ctx context.Context, c command.Cli, args ...string) error {
	var attempt uint8
	op := func() (bool, error) {
		return false, doCliPull(ctx, c, args...)
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

func doCliPush(ctx context.Context, c command.Cli, args ...string) error {
	return prepareCliCmd(ctx, image.NewPushCommand(c), args...).Execute()
}

const cliPushMaxAttempts uint8 = 10

func doCliPushWithRetries(ctx context.Context, c command.Cli, args ...string) error {
	var attempt uint8
	op := func() (bool, error) {
		err := doCliPush(ctx, c, args...)
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

func CliTag(ctx context.Context, args ...string) error {
	if len(args) < 2 {
		return fmt.Errorf("tag requires source and target arguments")
	}
	return apiCli(ctx).ImageTag(ctx, args[0], args[1])
}

func CliRmi(ctx context.Context, args ...string) error {
	force := false
	imageRefs := []string{}

	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else {
			imageRefs = append(imageRefs, arg)
		}
	}

	for _, ref := range imageRefs {
		_, err := apiCli(ctx).ImageRemove(ctx, ref, types.ImageRemoveOptions{
			Force:         force,
			PruneChildren: false,
		})
		if err != nil {
			return fmt.Errorf("remove image %s: %w", ref, err)
		}
	}

	return nil
}

type BuildOptions struct {
	EnableBuildx bool
}

func doCliBuild(ctx context.Context, c command.Cli, opts BuildOptions, args ...string) error {
	cmd := NewBuildxCommand(c)
	finalArgs := append([]string{"build", "--load"}, args...)
	return prepareCliCmd(ctx, cmd, finalArgs...).Execute()
}

func CliBuild_LiveOutputWithCustomIn(ctx context.Context, rc io.ReadCloser, args ...string) error {
	args = append([]string{"--provenance=false"}, args...)

	return cliWithCustomOptions(ctx, []command.DockerCliOption{
		func(cli *command.DockerCli) error {
			cli.SetIn(streams.NewIn(rc))
			return nil
		},
	}, func(cli command.Cli) error {
		return doCliBuild(ctx, cli, BuildOptions{}, args...)
	})
}

func CliImageSaveToStream(ctx context.Context, imageName string) (io.ReadCloser, error) {
	return apiCli(ctx).ImageSave(ctx, []string{imageName})
}

func CliLoadFromStream(ctx context.Context, input io.Reader) (string, error) {
	loadResponse, err := apiCli(ctx).ImageLoad(ctx, input, true)
	if err != nil {
		return "", fmt.Errorf("load failed: %w", err)
	}
	defer loadResponse.Body.Close()

	body, err := io.ReadAll(loadResponse.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return parseIDDigestFromImageLoadResponseBody(body), nil
}

func parseIDDigestFromImageLoadResponseBody(body []byte) string {
	// We always have a string of fixed length like bellow when use cli directly:
	// `Loaded image ID: sha256:26b2eb03618e749084668eaff68cff8f81dda12d06ac641be7a6398b82a6f25b`
	// Here we have json-wrapped representation of this string:
	// `{"stream":"Loaded image ID: sha256:26b2eb03618e749084668eaff68cff8f81dda12d06ac641be7a6398b82a6f25b\n"}\n`
	// So we can just slice it using these knowledges.

	// trim trailing \n
	bodySanitized := bytes.TrimSpace(body)

	n := len(bodySanitized) - len(`\n"}`) // json ending offset
	digestSize := 64
	digest := bodySanitized[n-digestSize : n]

	return string(digest)
}
