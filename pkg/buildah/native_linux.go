//go:build linux
// +build linux

package buildah

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/docker"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/buildah/pkg/parse"
	"github.com/containers/common/libimage"
	"github.com/containers/image/v5/manifest"
	imgstor "github.com/containers/image/v5/storage"
	"github.com/containers/image/v5/transports/alltransports"
	imgtypes "github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/containers/storage/drivers/overlay"
	"github.com/containers/storage/pkg/homedir"
	"github.com/containers/storage/pkg/reexec"
	"github.com/containers/storage/pkg/unshare"
	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-shellwords"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah/thirdparty"
)

const (
	MaxPullPushRetries = 3
	PullPushRetryDelay = 2 * time.Second
)

func NativeProcessStartupHook() bool {
	if reexec.Init() {
		return true
	}

	if debug() {
		logrus.SetLevel(logrus.TraceLevel)
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	return false
}

type NativeBuildah struct {
	BaseBuildah

	Store                     storage.Store
	Runtime                   libimage.Runtime
	DefaultSystemContext      imgtypes.SystemContext
	DefaultCommonBuildOptions define.CommonBuildOptions

	platforms []struct{ OS, Arch, Variant string }
}

func NewNativeBuildah(commonOpts CommonBuildahOpts, opts NativeModeOpts) (*NativeBuildah, error) {
	b := &NativeBuildah{}

	baseBuildah, err := NewBaseBuildah(commonOpts.TmpDir, BaseBuildahOpts{
		Isolation: *commonOpts.Isolation,
		Insecure:  commonOpts.Insecure,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create BaseBuildah: %w", err)
	}
	b.BaseBuildah = *baseBuildah

	storeOpts, err := NewNativeStoreOptions(unshare.GetRootlessUID(), *commonOpts.StorageDriver)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize storage opts: %w", err)
	}

	b.Store, err = storage.GetStore(storage.StoreOptions(*storeOpts))
	if err != nil {
		return nil, fmt.Errorf("unable to get storage: %w", err)
	}

	b.DefaultSystemContext = imgtypes.SystemContext{
		SignaturePolicyPath:               b.SignaturePolicyPath,
		SystemRegistriesConfPath:          b.RegistriesConfigPath,
		SystemRegistriesConfDirPath:       b.RegistriesConfigDirPath,
		OCIInsecureSkipTLSVerify:          b.Insecure,
		DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
		DockerDaemonInsecureSkipTLSVerify: b.Insecure,
	}

	if opts.Platform != "" {
		os, arch, variant, err := parse.Platform(opts.Platform)
		if err != nil {
			return nil, fmt.Errorf("unable to parse platform %q: %w", opts.Platform, err)
		}

		b.DefaultSystemContext.OSChoice = os
		b.DefaultSystemContext.ArchitectureChoice = arch
		b.DefaultSystemContext.VariantChoice = variant

		b.platforms = []struct{ OS, Arch, Variant string }{
			{os, arch, variant},
		}
	}

	b.DefaultCommonBuildOptions = define.CommonBuildOptions{
		ShmSize: DefaultShmSize,
	}

	imgstor.Transport.SetStore(b.Store)
	runtime, err := libimage.RuntimeFromStore(b.Store, &libimage.RuntimeOptions{
		SystemContext: &b.DefaultSystemContext,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting runtime from store: %w", err)
	}
	b.Runtime = *runtime

	return b, nil
}

// Inspect returns nil, nil if image not found.
func (b *NativeBuildah) Inspect(ctx context.Context, ref string) (*thirdparty.BuilderInfo, error) {
	builder, err := b.getBuilderFromImage(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("error doing inspect: %w", err)
	}
	if builder == nil {
		return nil, nil
	}

	buildInfo := thirdparty.BuilderInfo(buildah.GetBuildInfo(builder))

	return &buildInfo, nil
}

func (b *NativeBuildah) Tag(_ context.Context, ref, newRef string, opts TagOpts) error {
	image, err := b.getImage(ref)
	if err != nil {
		return err
	}

	if err := image.Tag(newRef); err != nil {
		return fmt.Errorf("error tagging image: %w", err)
	}

	return nil
}

func (b *NativeBuildah) Push(ctx context.Context, ref string, opts PushOpts) error {
	pushOpts := buildah.PushOptions{
		Compression:         define.Gzip,
		SignaturePolicyPath: b.SignaturePolicyPath,
		ReportWriter:        opts.LogWriter,
		Store:               b.Store,
		SystemContext:       &b.DefaultSystemContext,
		ManifestType:        manifest.DockerV2Schema2MediaType,
		MaxRetries:          MaxPullPushRetries,
		RetryDelay:          PullPushRetryDelay,
	}

	imageRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", ref))
	if err != nil {
		return fmt.Errorf("error parsing image ref from %q: %w", ref, err)
	}

	if _, _, err = buildah.Push(ctx, ref, imageRef, pushOpts); err != nil {
		return fmt.Errorf("error pushing image %q: %w", ref, err)
	}

	return nil
}

func (b *NativeBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	buildOpts := define.BuildOptions{
		Isolation:               define.Isolation(b.Isolation),
		Runtime:                 DefaultRuntime,
		Args:                    opts.BuildArgs,
		SignaturePolicyPath:     b.SignaturePolicyPath,
		ReportWriter:            opts.LogWriter,
		OutputFormat:            buildah.Dockerv2ImageManifest,
		SystemContext:           &b.DefaultSystemContext,
		ConfigureNetwork:        define.NetworkEnabled,
		CommonBuildOpts:         &b.DefaultCommonBuildOptions,
		Target:                  opts.Target,
		MaxPullPushRetries:      MaxPullPushRetries,
		PullPushRetryDelay:      PullPushRetryDelay,
		Platforms:               b.platforms,
		Layers:                  true,
		RemoveIntermediateCtrs:  true,
		ForceRmIntermediateCtrs: false,
		NoCache:                 false,
	}

	errLog := &bytes.Buffer{}
	if opts.LogWriter != nil {
		buildOpts.Out = opts.LogWriter
		buildOpts.Err = io.MultiWriter(opts.LogWriter, errLog)
	} else {
		buildOpts.Err = errLog
	}

	sessionTmpDir, contextTmpDir, dockerfileTmpPath, err := b.prepareBuildFromDockerfile(dockerfile, opts.ContextTar)
	if err != nil {
		return "", fmt.Errorf("error preparing for build from dockerfile: %w", err)
	}
	defer func() {
		if err = os.RemoveAll(sessionTmpDir); err != nil {
			logboek.Warn().LogF("unable to remove session tmp dir %q: %s\n", sessionTmpDir, err)
		}
	}()
	buildOpts.ContextDirectory = contextTmpDir

	imageId, _, err := imagebuildah.BuildDockerfiles(ctx, b.Store, buildOpts, dockerfileTmpPath)
	if err != nil {
		return "", fmt.Errorf("unable to build Dockerfile %q:\n%s\n%w", dockerfileTmpPath, errLog.String(), err)
	}

	return imageId, nil
}

func (b *NativeBuildah) Mount(ctx context.Context, container string, opts MountOpts) (string, error) {
	builder, err := b.openContainerBuilder(ctx, container)
	if err != nil {
		return "", fmt.Errorf("unable to open container %q builder: %w", container, err)
	}

	return builder.Mount("")
}

func (b *NativeBuildah) Umount(ctx context.Context, container string, opts UmountOpts) error {
	builder, err := b.openContainerBuilder(ctx, container)
	if err != nil {
		return fmt.Errorf("unable to open container %q builder: %w", container, err)
	}

	return builder.Unmount()
}

func (b *NativeBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	runOpts := buildah.RunOptions{
		Isolation:        define.Isolation(b.Isolation),
		Runtime:          DefaultRuntime,
		Args:             opts.Args,
		Mounts:           opts.Mounts,
		ConfigureNetwork: define.NetworkEnabled,
		SystemContext:    &b.DefaultSystemContext,
	}

	stderr := &bytes.Buffer{}
	if opts.LogWriter != nil {
		runOpts.Stdout = opts.LogWriter
		runOpts.Stderr = io.MultiWriter(opts.LogWriter, stderr)
	} else {
		runOpts.Stderr = stderr
	}

	builder, err := b.openContainerBuilder(ctx, container)
	if err != nil {
		return fmt.Errorf("unable to open container %q builder: %w", container, err)
	}

	if err := builder.Run(command, runOpts); err != nil {
		return fmt.Errorf("RunCommand failed:\n%s\n%w", stderr.String(), err)
	}

	return nil
}

func (b *NativeBuildah) FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) (string, error) {
	builder, err := buildah.NewBuilder(ctx, b.Store, buildah.BuilderOptions{
		FromImage:           image,
		Container:           container,
		SignaturePolicyPath: b.SignaturePolicyPath,
		ReportWriter:        opts.LogWriter,
		SystemContext:       &b.DefaultSystemContext,
		Isolation:           define.Isolation(b.Isolation),
		ConfigureNetwork:    define.NetworkEnabled,
		CommonBuildOpts:     &b.DefaultCommonBuildOptions,
		Format:              buildah.Dockerv2ImageManifest,
		MaxPullRetries:      MaxPullPushRetries,
		PullRetryDelay:      PullPushRetryDelay,
		Capabilities:        define.DefaultCapabilities,
	})
	if err != nil {
		return "", fmt.Errorf("unable to create builder: %w", err)
	}

	return builder.Container, builder.Save()
}

func (b *NativeBuildah) Pull(ctx context.Context, ref string, opts PullOpts) error {
	pullOpts := buildah.PullOptions{
		SignaturePolicyPath: b.SignaturePolicyPath,
		ReportWriter:        opts.LogWriter,
		Store:               b.Store,
		SystemContext:       &b.DefaultSystemContext,
		MaxRetries:          MaxPullPushRetries,
		RetryDelay:          PullPushRetryDelay,
		PullPolicy:          define.PullIfNewer,
	}

	imageID, err := buildah.Pull(ctx, ref, pullOpts)
	if err != nil {
		return fmt.Errorf("error pulling image %q: %w", ref, err)
	}

	imageInspect, err := b.Inspect(ctx, imageID)
	if err != nil {
		return fmt.Errorf("unable to inspect pulled image %q: %w", imageID, err)
	}

	platformMismatch := false
	if b.DefaultSystemContext.OSChoice != "" && b.DefaultSystemContext.OSChoice != imageInspect.OCIv1.OS {
		platformMismatch = true
	}
	if b.DefaultSystemContext.ArchitectureChoice != "" && b.DefaultSystemContext.ArchitectureChoice != imageInspect.OCIv1.Architecture {
		platformMismatch = true
	}
	if b.DefaultSystemContext.VariantChoice != "" && b.DefaultSystemContext.VariantChoice != imageInspect.OCIv1.Variant {
		platformMismatch = true
	}

	if platformMismatch {
		imagePlatform := fmt.Sprintf("%s/%s/%s", imageInspect.OCIv1.OS, imageInspect.OCIv1.Architecture, imageInspect.OCIv1.Variant)
		expectedPlatform := fmt.Sprintf("%s/%s", b.DefaultSystemContext.OSChoice, b.DefaultSystemContext.ArchitectureChoice)
		if b.DefaultSystemContext.VariantChoice != "" {
			expectedPlatform = fmt.Sprintf("%s/%s", expectedPlatform, b.DefaultSystemContext.VariantChoice)
		}
		return fmt.Errorf("image platform mismatch: image uses %s, expecting %s platform", imagePlatform, expectedPlatform)
	}

	return nil
}

func (b *NativeBuildah) Rm(ctx context.Context, ref string, opts RmOpts) error {
	builder, err := b.getBuilderFromContainer(ctx, ref)
	if err != nil {
		return fmt.Errorf("error getting builder: %w", err)
	}

	return builder.Delete()
}

func (b *NativeBuildah) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	_, rmiErrors := b.Runtime.RemoveImages(ctx, []string{ref}, &libimage.RemoveImagesOptions{
		Force: opts.Force,
		// Filters: []string{"readonly=false", "intermediate=false", "dangling=true"},
	})

	var multiErr *multierror.Error
	return multierror.Append(multiErr, rmiErrors...).ErrorOrNil()
}

func (b *NativeBuildah) Commit(ctx context.Context, container string, opts CommitOpts) (string, error) {
	builder, err := b.getBuilderFromContainer(ctx, container)
	if err != nil {
		return "", fmt.Errorf("error getting builder: %w", err)
	}

	var imageRef imgtypes.ImageReference
	if opts.Image != "" {
		imageRef, err = alltransports.ParseImageName(opts.Image)
		if err != nil {
			return "", fmt.Errorf("error parsing image name: %w", err)
		}
	}

	builder.SetLabel("werf.io/base-image-id", fmt.Sprintf("sha256:%s", builder.FromImageID))

	imgID, _, _, err := builder.Commit(ctx, imageRef, buildah.CommitOptions{
		PreferredManifestType: buildah.Dockerv2ImageManifest,
		SignaturePolicyPath:   b.SignaturePolicyPath,
		ReportWriter:          opts.LogWriter,
		SystemContext:         &b.DefaultSystemContext,
		MaxRetries:            MaxPullPushRetries,
		RetryDelay:            PullPushRetryDelay,
	})
	if err != nil {
		return "", fmt.Errorf("error doing commit: %w", err)
	}

	return imgID, nil
}

func (b *NativeBuildah) Config(ctx context.Context, container string, opts ConfigOpts) error {
	builder, err := b.getBuilderFromContainer(ctx, container)
	if err != nil {
		return fmt.Errorf("error getting builder: %w", err)
	}

	for _, label := range opts.Labels {
		labelSlice := strings.SplitN(label, "=", 2)
		switch {
		case len(labelSlice) > 1:
			builder.SetLabel(labelSlice[0], labelSlice[1])
		case labelSlice[0] == "-":
			builder.ClearLabels()
		case strings.HasSuffix(labelSlice[0], "-"):
			builder.UnsetLabel(strings.TrimSuffix(labelSlice[0], "-"))
		default:
			builder.SetLabel(labelSlice[0], "")
		}
	}

	for name, value := range opts.Envs {
		builder.SetEnv(name, value)
	}

	for _, volume := range opts.Volumes {
		builder.AddVolume(volume)
	}

	for _, expose := range opts.Expose {
		builder.SetPort(expose)
	}

	if len(opts.Cmd) > 0 {
		builder.SetCmd(opts.Cmd)
	}

	if len(opts.Entrypoint) > 0 {
		builder.SetEntrypoint(opts.Entrypoint)
	}

	if opts.User != "" {
		builder.SetUser(opts.User)
	}

	if opts.Workdir != "" {
		builder.SetWorkDir(opts.Workdir)
	}

	if opts.Healthcheck != "" {
		if healthcheck, err := newHealthConfigFromString(opts.Healthcheck); err != nil {
			return fmt.Errorf("error creating HEALTHCHECK: %w", err)
		} else if healthcheck != nil {
			builder.SetHealthcheck(healthcheck)
		}
	}

	return builder.Save()
}

func (b *NativeBuildah) getImage(ref string) (*libimage.Image, error) {
	image, _, err := b.Runtime.LookupImage(ref, &libimage.LookupImageOptions{
		ManifestList: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error looking up image %q: %w", ref, err)
	}

	return image, nil
}

// getBuilderFromImage returns nil, nil if image not found.
func (b *NativeBuildah) getBuilderFromImage(ctx context.Context, imgName string) (builder *buildah.Builder, err error) {
	builder, err = buildah.ImportBuilderFromImage(ctx, b.Store, buildah.ImportFromImageOptions{
		Image:               imgName,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext:       &b.DefaultSystemContext,
	})
	switch {
	case err != nil && strings.HasSuffix(err.Error(), storage.ErrImageUnknown.Error()):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("error getting builder from image %q: %w", imgName, err)
	case builder == nil:
		panic("error mocking up build configuration")
	}

	return builder, nil
}

func (b *NativeBuildah) getBuilderFromContainer(ctx context.Context, container string) (*buildah.Builder, error) {
	var builder *buildah.Builder
	var err error

	builder, err = buildah.OpenBuilder(b.Store, container)
	if os.IsNotExist(errors.Cause(err)) {
		builder, err = buildah.ImportBuilder(ctx, b.Store, buildah.ImportOptions{
			Container:           container,
			SignaturePolicyPath: b.SignaturePolicyPath,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("unable to open builder: %w", err)
	}
	if builder == nil {
		return nil, fmt.Errorf("error finding build container")
	}

	return builder, nil
}

func (b *NativeBuildah) openContainerBuilder(ctx context.Context, container string) (*buildah.Builder, error) {
	builder, err := buildah.OpenBuilder(b.Store, container)
	switch {
	case os.IsNotExist(errors.Cause(err)):
		builder, err = buildah.ImportBuilder(ctx, b.Store, buildah.ImportOptions{
			Container:           container,
			SignaturePolicyPath: b.SignaturePolicyPath,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to import builder for container %q: %w", container, err)
		}
	case err != nil:
		return nil, fmt.Errorf("unable to open builder for container %q: %w", container, err)
	}

	return builder, err
}

func NewNativeStoreOptions(rootlessUID int, driver StorageDriver) (*thirdparty.StoreOptions, error) {
	var (
		runRoot string
		err     error
	)

	if rootlessUID == 0 {
		runRoot = "/run/containers/storage"
	} else {
		runRoot, err = storage.GetRootlessRuntimeDir(rootlessUID)
		if err != nil {
			return nil, fmt.Errorf("unable to get runtime dir: %w", err)
		}
	}

	home, err := homedir.GetDataHome()
	if err != nil {
		return nil, fmt.Errorf("unable to get HOME data dir: %w", err)
	}

	rootlessStoragePath := filepath.Join(home, "containers", "storage")

	var graphRoot string
	if rootlessUID == 0 {
		graphRoot = "/var/lib/containers/storage"
	} else {
		graphRoot = rootlessStoragePath
	}

	var graphDriverOptions []string
	if driver == StorageDriverOverlay {
		supportsNative, err := overlay.SupportsNativeOverlay(graphRoot, runRoot)
		if err != nil {
			return nil, fmt.Errorf("unable to check native overlayfs support: %w", err)
		}

		if !supportsNative {
			fuseOpts, err := GetFuseOverlayfsOptions()
			if err != nil {
				return nil, fmt.Errorf("unable to get fuse overlayfs options: %w", err)
			}

			graphDriverOptions = append(graphDriverOptions, fuseOpts...)
		}
	}

	return &thirdparty.StoreOptions{
		RunRoot:             runRoot,
		GraphRoot:           graphRoot,
		RootlessStoragePath: rootlessStoragePath,
		GraphDriverName:     string(driver),
		GraphDriverOptions:  graphDriverOptions,
	}, nil
}

// Can return nil pointer to HealthConfig.
func newHealthConfigFromString(healthcheck string) (*docker.HealthConfig, error) {
	if healthcheck == "" {
		return nil, nil
	}

	healthConfig := &docker.HealthConfig{}

	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			switch args[0] {
			case "NONE":
				healthConfig = &docker.HealthConfig{
					Test: []string{"NONE"},
				}
				return nil
			case "CMD", "CMD-SHELL":
				if len(args) == 1 {
					return fmt.Errorf("HEALTHCHECK %s should have command specified", args[0])
				}
				healthConfig.Test = args
				return nil
			}

			healthConfig = nil
			return nil
		},
	}

	flags := cmd.Flags()
	flags.DurationVar(&healthConfig.Interval, "interval", 30*time.Second, "")
	flags.DurationVar(&healthConfig.Timeout, "timeout", 30*time.Second, "")
	flags.DurationVar(&healthConfig.StartPeriod, "start-period", 0, "")
	flags.IntVar(&healthConfig.Retries, "retries", 3, "")

	healthcheckSlice, err := shellwords.Parse(healthcheck)
	if err != nil {
		return nil, fmt.Errorf("error parsing HEALTHCHECK: %w", err)
	}
	cmd.SetArgs(healthcheckSlice)

	if err := cmd.Execute(); err != nil {
		return nil, fmt.Errorf("error parsing HEALTHCHECK: %w", err)
	}

	return healthConfig, nil
}
