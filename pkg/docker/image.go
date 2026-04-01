package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/distribution/reference"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	registryTypes "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	dockerbuildkit "github.com/docker/docker/client/buildkit"
	"github.com/docker/docker/pkg/jsonmessage"
	buildkitclient "github.com/moby/buildkit/client"
	buildkitexptypes "github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
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

func parseImageRef(imageRef string) (reference.Named, error) {
	ref, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func getRegistryAuth(ctx context.Context, imageRef string) (string, error) {
	ref, err := parseImageRef(imageRef)
	if err != nil {
		return "", fmt.Errorf("parse image ref: %w", err)
	}

	hostname := reference.Domain(ref)

	authConfig := command.ResolveAuthConfig(cli(ctx).ConfigFile(), &registryTypes.IndexInfo{Name: hostname})

	encodedAuth, err := registryTypes.EncodeAuthConfig(authConfig)
	if err != nil {
		return "", fmt.Errorf("encode auth config: %w", err)
	}

	return encodedAuth, nil
}

func doCliPull(ctx context.Context, args ...string) error {
	var platform string
	var imageRef string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--platform" && i+1 < len(args) {
			platform = args[i+1]
			i++
		} else {
			imageRef = arg
		}
	}

	if imageRef == "" {
		return fmt.Errorf("image reference required")
	}

	registryAuth, err := getRegistryAuth(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("get registry auth: %w", err)
	}

	pullResp, err := apiCli(ctx).ImagePull(ctx, imageRef, types.ImagePullOptions{
		Platform:     platform,
		RegistryAuth: registryAuth,
	})
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	defer pullResp.Close()

	_, err = io.Copy(io.Discard, pullResp)
	if err != nil {
		return fmt.Errorf("read pull response: %w", err)
	}

	return nil
}

func CliPull(ctx context.Context, args ...string) error {
	return doCliPull(ctx, args...)
}

const cliPullMaxAttempts uint8 = 5

func doCliPullWithRetries(ctx context.Context, args ...string) error {
	var attempt uint8
	op := func() (bool, error) {
		return false, doCliPull(ctx, args...)
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
	return doCliPullWithRetries(ctx, args...)
}

func doCliPush(ctx context.Context, args ...string) error {
	var imageRef string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			imageRef = arg
		}
	}

	if imageRef == "" {
		return fmt.Errorf("image reference required")
	}

	registryAuth, err := getRegistryAuth(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("get registry auth: %w", err)
	}

	pushResp, err := apiCli(ctx).ImagePush(ctx, imageRef, types.ImagePushOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		return fmt.Errorf("push image: %w", err)
	}
	defer pushResp.Close()

	err = jsonmessage.DisplayJSONMessagesStream(pushResp, io.Discard, 0, false, nil)
	if err != nil {
		return fmt.Errorf("push image: %w", err)
	}

	return nil
}

const cliPushMaxAttempts uint8 = 10

func doCliPushWithRetries(ctx context.Context, args ...string) error {
	var attempt uint8
	op := func() (bool, error) {
		err := doCliPush(ctx, args...)
		return false, err
	}
	notify := func(err error, duration time.Duration) {
		logboek.Context(ctx).Warn().LogF("Retrying docker push in %0.2f seconds (%d/%d) ...\n", duration.Seconds(), attempt, cliPushMaxAttempts)
	}
	return doCliOperationWithRetries(ctx, op, &attempt, cliPushMaxAttempts, notify)
}

func CliPushWithRetries(ctx context.Context, args ...string) error {
	return doCliPushWithRetries(ctx, args...)
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

func CliBuild_LiveOutputWithCustomIn(ctx context.Context, rc io.ReadCloser, args ...string) error {
	args = append([]string{"--provenance=false"}, args...)
	return doCliBuildSDK(ctx, rc, args...)
}

func doCliBuildSDK(ctx context.Context, inStream io.ReadCloser, args ...string) error {
	ctx = ensureLogboekContext(ctx)

	var (
		dockerfilePath string
		platform       string
		target         string
		networkMode    string
		provenance     string
		metadataFile   string
		addHosts       []string
		buildArgs      = map[string]string{}
		labels         = map[string]string{}
		secrets        []secretsprovider.Source
		sshConfigs     []sshprovider.AgentConfig
		tags           []string
		useStdin       bool
		setProvenance  bool
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-":
			useStdin = true
		case arg == "--file" || arg == "-f":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			dockerfilePath = value
		case strings.HasPrefix(arg, "--file="):
			dockerfilePath = strings.TrimPrefix(arg, "--file=")
		case arg == "--platform":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			platform = value
		case strings.HasPrefix(arg, "--platform="):
			platform = strings.TrimPrefix(arg, "--platform=")
		case arg == "--target":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			target = value
		case strings.HasPrefix(arg, "--target="):
			target = strings.TrimPrefix(arg, "--target=")
		case arg == "--network":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			networkMode = value
		case strings.HasPrefix(arg, "--network="):
			networkMode = strings.TrimPrefix(arg, "--network=")
		case arg == "--ssh":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			config, err := parseSSHConfig(value)
			if err != nil {
				return err
			}
			sshConfigs = append(sshConfigs, config)
		case strings.HasPrefix(arg, "--ssh="):
			value := strings.TrimPrefix(arg, "--ssh=")
			config, err := parseSSHConfig(value)
			if err != nil {
				return err
			}
			sshConfigs = append(sshConfigs, config)
		case arg == "--add-host":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			addHosts = append(addHosts, value)
		case strings.HasPrefix(arg, "--add-host="):
			addHosts = append(addHosts, strings.TrimPrefix(arg, "--add-host="))
		case arg == "--build-arg":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			key, val, err := parseKeyValue(value)
			if err != nil {
				return err
			}
			buildArgs[key] = val
		case strings.HasPrefix(arg, "--build-arg="):
			value := strings.TrimPrefix(arg, "--build-arg=")
			key, val, err := parseKeyValue(value)
			if err != nil {
				return err
			}
			buildArgs[key] = val
		case arg == "--label":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			key, val, err := parseKeyValue(value)
			if err != nil {
				return err
			}
			labels[key] = val
		case strings.HasPrefix(arg, "--label="):
			value := strings.TrimPrefix(arg, "--label=")
			key, val, err := parseKeyValue(value)
			if err != nil {
				return err
			}
			labels[key] = val
		case arg == "--secret":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			secret, err := parseSecret(value)
			if err != nil {
				return err
			}
			secrets = append(secrets, secret)
		case strings.HasPrefix(arg, "--secret="):
			value := strings.TrimPrefix(arg, "--secret=")
			secret, err := parseSecret(value)
			if err != nil {
				return err
			}
			secrets = append(secrets, secret)
		case arg == "--tag":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			tags = append(tags, value)
		case strings.HasPrefix(arg, "--tag="):
			tags = append(tags, strings.TrimPrefix(arg, "--tag="))
		case arg == "--metadata-file":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			metadataFile = value
		case strings.HasPrefix(arg, "--metadata-file="):
			metadataFile = strings.TrimPrefix(arg, "--metadata-file=")
		case arg == "--provenance":
			value, err := getFlagValue(args, &i, arg)
			if err != nil {
				return err
			}
			provenance = value
			setProvenance = true
		case strings.HasPrefix(arg, "--provenance="):
			provenance = strings.TrimPrefix(arg, "--provenance=")
			setProvenance = true
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unsupported docker build flag %q", arg)
		default:
			return fmt.Errorf("unsupported build argument %q", arg)
		}
	}

	if !useStdin {
		return fmt.Errorf("build context must be provided via '-' argument")
	}
	if inStream == nil {
		return fmt.Errorf("build context stream is nil")
	}

	contextDir, err := extractBuildContext(inStream)
	if err != nil {
		return fmt.Errorf("extract build context: %w", err)
	}
	defer os.RemoveAll(contextDir)

	frontendAttrs := map[string]string{}
	if dockerfilePath != "" {
		frontendAttrs["filename"] = dockerfilePath
	}
	if platform != "" {
		frontendAttrs["platform"] = platform
	}
	if target != "" {
		frontendAttrs["target"] = target
	}
	if networkMode != "" {
		frontendAttrs["force-network-mode"] = networkMode
	}
	if len(addHosts) > 0 {
		frontendAttrs["add-hosts"] = strings.Join(addHosts, ",")
	}
	for key, value := range buildArgs {
		frontendAttrs["build-arg:"+key] = value
	}
	for key, value := range labels {
		frontendAttrs["label:"+key] = value
	}
	if setProvenance {
		if strings.EqualFold(provenance, "false") {
			frontendAttrs["attest:provenance"] = ""
		} else if provenance != "" {
			frontendAttrs["attest:provenance"] = provenance
		}
	}

	solveOpt := buildkitclient.SolveOpt{
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		LocalDirs: map[string]string{
			"context":    contextDir,
			"dockerfile": contextDir,
		},
	}

	if len(secrets) > 0 {
		store, err := secretsprovider.NewStore(secrets)
		if err != nil {
			return fmt.Errorf("create secrets store: %w", err)
		}
		solveOpt.Session = append(solveOpt.Session, secretsprovider.NewSecretProvider(store))
	}
	if len(sshConfigs) > 0 {
		provider, err := sshprovider.NewSSHAgentProvider(sshConfigs)
		if err != nil {
			return fmt.Errorf("create ssh provider: %w", err)
		}
		solveOpt.Session = append(solveOpt.Session, provider)
	}

	exportAttrs := map[string]string{}
	if len(tags) > 0 {
		exportAttrs[string(buildkitexptypes.OptKeyName)] = strings.Join(tags, ",")
	}
	solveOpt.Exports = []buildkitclient.ExportEntry{
		{
			Type:  "moby",
			Attrs: exportAttrs,
		},
	}

	bk, err := buildkitclient.New(ctx, "", dockerbuildkit.ClientOpts(apiCli(ctx))...)
	if err != nil {
		return fmt.Errorf("create buildkit client: %w", err)
	}
	defer bk.Close()

	statusCh := make(chan *buildkitclient.SolveStatus)
	go func() {
		for range statusCh {
		}
	}()

	resp, err := bk.Solve(ctx, nil, solveOpt, statusCh)
	if err != nil {
		return fmt.Errorf("solve build: %w", err)
	}

	if metadataFile != "" {
		if resp == nil {
			return fmt.Errorf("build response is nil")
		}
		digest := resp.ExporterResponse["containerimage.digest"]
		if digest == "" {
			return fmt.Errorf("containerimage.digest not found in build response")
		}
		payload, err := json.Marshal(map[string]string{"containerimage.digest": digest})
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(metadataFile), 0o755); err != nil {
			return fmt.Errorf("create metadata directory: %w", err)
		}
		if err := os.WriteFile(metadataFile, payload, 0o644); err != nil {
			return fmt.Errorf("write metadata file: %w", err)
		}
	}

	return nil
}

func extractBuildContext(inStream io.Reader) (string, error) {
	contextDir, err := os.MkdirTemp("", "werf-buildkit-context-*")
	if err != nil {
		return "", err
	}
	cleanup := func() {
		_ = os.RemoveAll(contextDir)
	}

	tr := tar.NewReader(inStream)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			cleanup()
			return "", err
		}
		if hdr == nil {
			continue
		}

		target, err := safeJoin(contextDir, hdr.Name)
		if err != nil {
			cleanup()
			return "", err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				cleanup()
				return "", err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				cleanup()
				return "", err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				cleanup()
				return "", err
			}
			if _, err := io.Copy(file, tr); err != nil {
				_ = file.Close()
				cleanup()
				return "", err
			}
			if err := file.Close(); err != nil {
				cleanup()
				return "", err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				cleanup()
				return "", err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				cleanup()
				return "", err
			}
		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				cleanup()
				return "", err
			}
			linkTarget, err := safeJoin(contextDir, hdr.Linkname)
			if err != nil {
				cleanup()
				return "", err
			}
			if err := os.Link(linkTarget, target); err != nil {
				cleanup()
				return "", err
			}
		case tar.TypeXGlobalHeader, tar.TypeXHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
			continue
		default:
			cleanup()
			return "", fmt.Errorf("unsupported tar entry %q", hdr.Name)
		}
	}

	return contextDir, nil
}

func safeJoin(base, name string) (string, error) {
	cleanName := filepath.Clean(name)
	cleanName = strings.TrimPrefix(cleanName, string(filepath.Separator))
	baseClean := filepath.Clean(base)
	joined := filepath.Join(baseClean, cleanName)
	if joined != baseClean && !strings.HasPrefix(joined, baseClean+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid tar path %q", name)
	}
	return joined, nil
}

func getFlagValue(args []string, idx *int, flag string) (string, error) {
	if *idx+1 >= len(args) {
		return "", fmt.Errorf("flag %s requires value", flag)
	}
	(*idx)++
	return args[*idx], nil
}

func parseKeyValue(value string) (string, string, error) {
	key, val, found := strings.Cut(value, "=")
	if key == "" {
		return "", "", fmt.Errorf("invalid value %q", value)
	}
	if !found {
		return key, "", nil
	}
	return key, val, nil
}

func parseSecret(value string) (secretsprovider.Source, error) {
	var src secretsprovider.Source
	for _, part := range strings.Split(value, ",") {
		if part == "" {
			continue
		}
		key, val, found := strings.Cut(part, "=")
		if !found {
			return secretsprovider.Source{}, fmt.Errorf("invalid secret value %q", value)
		}
		switch key {
		case "id":
			src.ID = val
		case "src":
			src.FilePath = val
		case "env":
			src.Env = val
		default:
			return secretsprovider.Source{}, fmt.Errorf("unsupported secret option %q", key)
		}
	}
	if src.ID == "" {
		return secretsprovider.Source{}, fmt.Errorf("secret id is required")
	}
	return src, nil
}

func parseSSHConfig(value string) (sshprovider.AgentConfig, error) {
	if value == "" {
		return sshprovider.AgentConfig{}, fmt.Errorf("ssh config is empty")
	}
	key, val, found := strings.Cut(value, "=")
	if !found {
		return sshprovider.AgentConfig{ID: key}, nil
	}
	if key == "" {
		return sshprovider.AgentConfig{}, fmt.Errorf("ssh id is required")
	}
	paths := []string{}
	if val != "" {
		paths = strings.Split(val, ",")
	}
	return sshprovider.AgentConfig{ID: key, Paths: paths}, nil
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
