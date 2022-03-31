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
	"github.com/containers/buildah/imagebuildah"
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
	"github.com/sirupsen/logrus"
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

	Store                storage.Store
	Runtime              libimage.Runtime
	DefaultSystemContext imgtypes.SystemContext
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
		OCIInsecureSkipTLSVerify:          b.Insecure,
		DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
		DockerDaemonInsecureSkipTLSVerify: b.Insecure,
		SystemRegistriesConfPath:          b.RegistriesConfigPath,
		SystemRegistriesConfDirPath:       b.RegistriesConfigDirPath,
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
		Store:               b.Store,
		ManifestType:        manifest.DockerV2Schema2MediaType,
		MaxRetries:          MaxPullPushRetries,
		RetryDelay:          PullPushRetryDelay,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext:       &b.DefaultSystemContext,
	}

	if opts.LogWriter != nil {
		pushOpts.ReportWriter = opts.LogWriter
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
		Isolation:    define.Isolation(b.Isolation),
		OutputFormat: buildah.Dockerv2ImageManifest,
		CommonBuildOpts: &define.CommonBuildOptions{
			ShmSize: DefaultShmSize,
		},
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext:       &b.DefaultSystemContext,
		Args:                opts.BuildArgs,
		Target:              opts.Target,
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
		Args:      opts.Args,
		Isolation: define.Isolation(b.Isolation),
		Mounts:    opts.Mounts,
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
		FromImage: image,
		Container: container,
	})
	if err != nil {
		return "", fmt.Errorf("unable to create builder: %w", err)
	}

	return builder.Container, builder.Save()
}

func (b *NativeBuildah) Pull(ctx context.Context, ref string, opts PullOpts) error {
	pullOpts := buildah.PullOptions{
		Store:               b.Store,
		MaxRetries:          MaxPullPushRetries,
		RetryDelay:          PullPushRetryDelay,
		PullPolicy:          define.PullIfNewer,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext:       &b.DefaultSystemContext,
	}

	if opts.LogWriter != nil {
		pullOpts.ReportWriter = opts.LogWriter
	}

	if _, err := buildah.Pull(ctx, ref, pullOpts); err != nil {
		return fmt.Errorf("error pulling image %q: %w", ref, err)
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
		Filters: []string{"readonly=false", "intermediate=false", "dangling=true"},
		Force:   opts.Force,
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
			Container: container,
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
			Container: container,
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
