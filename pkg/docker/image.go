package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/docker/buildx/commands"
	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"
	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/confutil"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/samber/lo"
	"golang.org/x/net/context"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
)

type CreateImageOptions struct {
	Labels         []string
	TargetPlatform string
}

func CreateImage(ctx context.Context, ref string, opts CreateImageOptions) error {
	var importOpts dockerImage.ImportOptions
	if len(opts.Labels) > 0 {
		changeOption := "LABEL"
		for _, label := range opts.Labels {
			changeOption += fmt.Sprintf(" %s", label)
		}
		importOpts.Changes = append(importOpts.Changes, changeOption)
	}
	if opts.TargetPlatform != "" {
		importOpts.Platform = opts.TargetPlatform
	}
	_, err := apiCli(ctx).ImageImport(ctx, dockerImage.ImportSource{SourceName: "-"}, ref, importOpts)
	return err
}

func Images(ctx context.Context, options dockerImage.ListOptions) ([]dockerImage.Summary, error) {
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

func ImageInspect(ctx context.Context, ref string) (*dockerImage.InspectResponse, error) {
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
	cmd, err := lookupCliCommand(c, "pull")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
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
	cmd, err := lookupCliCommand(c, "push")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
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

func doCliTag(ctx context.Context, c command.Cli, args ...string) error {
	cmd, err := lookupCliCommand(c, "tag")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
}

func CliTag(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliTag(ctx, c, args...)
	})
}

func doCliRmi(ctx context.Context, c command.Cli, args ...string) error {
	cmd, err := lookupCliCommand(c, "rmi")
	if err != nil {
		return err
	}
	return prepareCliCmd(ctx, cmd, args...).Execute()
}

func CliRmi(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliRmi(ctx, c, args...)
	})
}

func CliRmi_LiveOutput(ctx context.Context, args ...string) error {
	return doCliRmi(ctx, cli(ctx), args...)
}

type CliBuildOptions struct {
	DockerfileName string
	Tags           []string
	BuildArgs      []string
	Labels         []string
	Target         string
	Platforms      []string
	Network        string
	ExtraHosts     []string
	SSH            string
	Secrets        []string
}

func CliBuild_LiveOutputWithCustomIn(ctx context.Context, rc io.ReadCloser, cliOpts CliBuildOptions) (string, error) {
	buildOpts := &commands.BuildOptions{
		ContextPath:            "-",
		ExportLoad:             true,
		DockerfileName:         cliOpts.DockerfileName,
		Tags:                   cliOpts.Tags,
		Target:                 cliOpts.Target,
		Platforms:              cliOpts.Platforms,
		NetworkMode:            cliOpts.Network,
		ExtraHosts:             cliOpts.ExtraHosts,
		ProvenanceResponseMode: string(confutil.MetadataProvenanceModeDisabled),
	}

	if len(cliOpts.BuildArgs) > 0 {
		buildOpts.BuildArgs = make(map[string]string, len(cliOpts.BuildArgs))
		for _, arg := range cliOpts.BuildArgs {
			k, v, _ := strings.Cut(arg, "=")
			buildOpts.BuildArgs[k] = v
		}
	}

	if len(cliOpts.Labels) > 0 {
		buildOpts.Labels = make(map[string]string, len(cliOpts.Labels))
		for _, label := range cliOpts.Labels {
			k, v, _ := strings.Cut(label, "=")
			buildOpts.Labels[k] = v
		}
	}

	if cliOpts.SSH != "" {
		sshSpecs, err := buildflags.ParseSSHSpecs([]string{cliOpts.SSH})
		if err != nil {
			return "", fmt.Errorf("parse ssh specs: %w", err)
		}
		buildOpts.SSH = sshSpecs
	}

	if len(cliOpts.Secrets) > 0 {
		secrets, err := buildflags.ParseSecretSpecs(cliOpts.Secrets)
		if err != nil {
			return "", fmt.Errorf("parse secret specs: %w", err)
		}
		buildOpts.Secrets = secrets
	}

	// TODO: properly handle index manifests instead of disabling provenance.
	// Provenance attestations create index manifests that werf cannot handle correctly.
	prevNoAttest, hadNoAttest := os.LookupEnv("BUILDX_NO_DEFAULT_ATTESTATIONS")
	os.Setenv("BUILDX_NO_DEFAULT_ATTESTATIONS", "1")
	defer func() {
		if hadNoAttest {
			os.Setenv("BUILDX_NO_DEFAULT_ATTESTATIONS", prevNoAttest)
		} else {
			os.Unsetenv("BUILDX_NO_DEFAULT_ATTESTATIONS")
		}
	}()

	progressMode := progressui.PlainMode
	if liveCliOutputEnabled {
		progressMode = progressui.AutoMode
	}

	dockerCli := cli(ctx)

	printer, err := progress.NewPrinter(ctx, logboek.Context(ctx).OutStream(), progressMode)
	if err != nil {
		return "", fmt.Errorf("create progress printer: %w", err)
	}

	resp, _, err := commands.RunBuild(ctx, dockerCli, buildOpts, rc, printer, nil)

	printErr := printer.Wait()
	if err == nil {
		err = printErr
	}
	if err != nil {
		return "", err
	}

	imageID := resp.ExporterResponse[exptypes.ExporterImageDigestKey]
	if imageID == "" {
		imageID = resp.ExporterResponse[exptypes.ExporterImageConfigDigestKey]
	}

	return imageID, nil
}

func CliImageSaveToStream(ctx context.Context, imageName string) (io.ReadCloser, error) {
	return apiCli(ctx).ImageSave(ctx, []string{imageName})
}

func CliLoadFromStream(ctx context.Context, input io.Reader) (string, error) {
	loadResponse, err := apiCli(ctx).ImageLoad(ctx, input)
	if err != nil {
		return "", fmt.Errorf("load failed: %w", err)
	}
	defer loadResponse.Body.Close()

	decoder := json.NewDecoder(loadResponse.Body)
	for {
		var msg jsonmessage.JSONMessage
		if err := decoder.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("decode load response: %w", err)
		}

		msg.Stream = strings.TrimSpace(msg.Stream)

		if _, imageID, hasID := strings.Cut(msg.Stream, "Loaded image ID: "); hasID {
			imageID = strings.TrimPrefix(imageID, "sha256:")
			return imageID, nil
		}

		if _, imageRef, hasRef := strings.Cut(msg.Stream, "Loaded image: "); hasRef {
			inspect, err := ImageInspect(ctx, imageRef)
			if err != nil {
				return "", fmt.Errorf("inspect loaded image %q: %w", imageRef, err)
			}
			return strings.TrimPrefix(inspect.ID, "sha256:"), nil
		}
	}

	return "", fmt.Errorf("no image ID or reference found in load response")
}
