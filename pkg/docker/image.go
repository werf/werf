package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/containerd/containerd/platforms"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/buildx/commands"
	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"
	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/confutil"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/util/progress/progressui"
	dockerImage "github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/samber/lo"
	"golang.org/x/net/context"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/container_backend/filter"
	"github.com/werf/werf/v3/pkg/container_backend/prune"
	"github.com/werf/werf/v3/pkg/opstats"
)

type CreateImageOptions struct {
	Labels         []string
	TargetPlatform string
}

func CreateImage(ctx context.Context, ref string, opts CreateImageOptions) error {
	defer opstats.Observe(ctx, "docker: image import")()

	var importOpts client.ImageImportOptions
	if len(opts.Labels) > 0 {
		changeOption := "LABEL"
		for _, label := range opts.Labels {
			changeOption += fmt.Sprintf(" %s", label)
		}
		importOpts.Changes = append(importOpts.Changes, changeOption)
	}
	if opts.TargetPlatform != "" {
		platform, err := platforms.Parse(opts.TargetPlatform)
		if err != nil {
			return fmt.Errorf("parse target platform %q: %w", opts.TargetPlatform, err)
		}
		importOpts.Platform = platform
	}

	// The Docker daemon reads the request body as the rootfs tar when fromSrc is "-".
	// A nil/empty body makes the daemon register a malformed empty layer that some
	// consumers (e.g. `dive`, `docker save` readers) fail to parse on Linux, so we
	// send a valid empty tar archive instead. This yields a single empty rootfs
	// layer, which is fine for `from: scratch`: later stapel stages copy imports on
	// top of it, and the empty layer is deterministic so it does not shift caching.
	emptyTar, err := emptyTarArchive()
	if err != nil {
		return fmt.Errorf("unable to build empty rootfs tar for image %q: %w", ref, err)
	}

	resp, err := apiCli(ctx).ImageImport(ctx, client.ImageImportSource{Source: emptyTar, SourceName: "-"}, ref, importOpts)
	if err != nil {
		return fmt.Errorf("unable to import image %q: %w", ref, err)
	}
	defer resp.Close()

	// Drain the response body so the daemon fully finalizes image creation before we return.
	if _, err := io.Copy(io.Discard, resp); err != nil {
		return fmt.Errorf("unable to read import response for image %q: %w", ref, err)
	}

	return nil
}

// emptyTarArchive returns a reader over a valid, empty tar archive (containing
// only the end-of-archive marker). It is used as the rootfs body for scratch
// images created via the Docker Engine import API.
func emptyTarArchive() (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("unable to close tar writer: %w", err)
	}
	return &buf, nil
}

func Images(ctx context.Context, options client.ImageListOptions) ([]dockerImage.Summary, error) {
	defer opstats.Observe(ctx, opstats.OperationDockerImageList)()

	result, err := apiCli(ctx).ImageList(ctx, options)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

func ImageExist(ctx context.Context, ref string) (bool, error) {
	if _, err := ImageInspect(ctx, ref); err != nil {
		if cerrdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ImageInspect(ctx context.Context, ref string) (*dockerImage.InspectResponse, error) {
	defer opstats.Observe(ctx, "docker: image inspect")()

	result, err := apiCli(ctx).ImageInspect(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &result.InspectResponse, nil
}

type (
	ImagesPruneOptions prune.Options
	ImagesPruneReport  prune.Report
)

// ImagesPrune containers using opts.Filters.
// List of accepted filters is there https://github.com/moby/moby/blob/25.0/daemon/containerd/image_prune.go#L22
func ImagesPrune(ctx context.Context, opts ImagesPruneOptions) (ImagesPruneReport, error) {
	defer opstats.Observe(ctx, "docker: image prune")()

	result, err := apiCli(ctx).ImagePrune(ctx, client.ImagePruneOptions{
		Filters: mapBackendFiltersToImagesPruneFilters(opts.Filters),
	})
	if err != nil {
		return ImagesPruneReport{}, err
	}
	itemsDeleted := lo.Map(result.Report.ImagesDeleted, func(item dockerImage.DeleteResponse, _ int) string {
		return item.Deleted
	})
	return ImagesPruneReport{
		ItemsDeleted:   itemsDeleted,
		SpaceReclaimed: result.Report.SpaceReclaimed,
	}, err
}

func mapBackendFiltersToImagesPruneFilters(list filter.FilterList) client.Filters {
	filters := make(client.Filters, len(list))
	for _, item := range list {
		filters.Add(item.First, item.Second)
	}
	return filters
}

func doCliPull(ctx context.Context, c command.Cli, args ...string) error {
	defer opstats.Observe(ctx, "docker: image pull")()

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
	defer opstats.Observe(ctx, "docker: image push")()

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
	defer opstats.Observe(ctx, "docker: image tag")()

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
	defer opstats.Observe(ctx, "docker: image remove")()

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

type CliBuildOptions struct {
	ContextPath    string
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
	defer opstats.Observe(ctx, "docker: image build")()

	buildOpts := &commands.BuildOptions{
		ContextPath:            cliOpts.ContextPath,
		ExportLoad:             true,
		DockerfileName:         cliOpts.DockerfileName,
		Tags:                   cliOpts.Tags,
		Target:                 cliOpts.Target,
		Platforms:              cliOpts.Platforms,
		NetworkMode:            cliOpts.Network,
		ExtraHosts:             cliOpts.ExtraHosts,
		ProvenanceResponseMode: string(confutil.MetadataProvenanceModeDisabled),
		Attests:                buildflags.Attests{{Type: "provenance", Disabled: true}},
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
	done := opstats.Observe(ctx, "docker: image save")

	rc, err := apiCli(ctx).ImageSave(ctx, []string{imageName})
	if err != nil {
		done()
		return nil, err
	}

	return opstats.NewObservedReadCloser(rc, done), nil
}

func CliLoadFromStream(ctx context.Context, input io.Reader) (string, error) {
	defer opstats.Observe(ctx, "docker: image load")()

	loadResponse, err := apiCli(ctx).ImageLoad(ctx, input)
	if err != nil {
		return "", fmt.Errorf("load failed: %w", err)
	}
	defer loadResponse.Close()

	decoder := json.NewDecoder(loadResponse)
	for {
		var msg struct {
			jsonstream.Message
			ErrorMessage string `json:"error,omitempty"`
		}
		if err := decoder.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("decode load response: %w", err)
		}

		if msg.Error != nil {
			return "", fmt.Errorf("load failed: %w", msg.Error)
		}
		if msg.ErrorMessage != "" {
			return "", fmt.Errorf("load failed: %s", msg.ErrorMessage)
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
