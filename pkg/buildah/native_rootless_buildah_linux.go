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
	"github.com/containers/storage/pkg/homedir"
	"github.com/containers/storage/pkg/reexec"
	"github.com/containers/storage/pkg/unshare"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah/types"
)

const (
	MaxPullPushRetries = 3
	PullPushRetryDelay = 2 * time.Second
)

func NativeRootlessProcessStartupHook() bool {
	if reexec.Init() {
		return true
	}

	if debug() {
		logrus.SetLevel(logrus.TraceLevel)
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	return false
}

type NativeRootlessBuildah struct {
	BaseBuildah

	Store   storage.Store
	Runtime libimage.Runtime
}

func NewNativeRootlessBuildah(commonOpts CommonBuildahOpts, opts NativeRootlessModeOpts) (*NativeRootlessBuildah, error) {
	b := &NativeRootlessBuildah{}

	baseBuildah, err := NewBaseBuildah(commonOpts.TmpDir, BaseBuildahOpts{
		Isolation: *commonOpts.Isolation,
		Insecure:  commonOpts.Insecure,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create BaseBuildah: %s", err)
	}
	b.BaseBuildah = *baseBuildah

	storeOpts, err := NewNativeStoreOptions(unshare.GetRootlessUID(), *commonOpts.StorageDriver)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize storage opts: %s", err)
	}

	b.Store, err = storage.GetStore(storage.StoreOptions(*storeOpts))
	if err != nil {
		return nil, fmt.Errorf("unable to get storage: %s", err)
	}

	imgstor.Transport.SetStore(b.Store)
	runtime, err := libimage.RuntimeFromStore(b.Store, &libimage.RuntimeOptions{
		SystemContext: &imgtypes.SystemContext{
			OCIInsecureSkipTLSVerify:          b.Insecure,
			DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
			DockerDaemonInsecureSkipTLSVerify: b.Insecure,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error getting runtime from store: %s", err)
	}
	b.Runtime = *runtime

	return b, nil
}

// Inspect returns nil, nil if image not found.
func (b *NativeRootlessBuildah) Inspect(ctx context.Context, ref string) (*types.BuilderInfo, error) {
	builder, err := b.getImageBuilder(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("error doing inspect: %s", err)
	}
	if builder == nil {
		return nil, nil
	}

	buildInfo := types.BuilderInfo(buildah.GetBuildInfo(builder))

	return &buildInfo, nil
}

func (b *NativeRootlessBuildah) Tag(_ context.Context, ref, newRef string, opts TagOpts) error {
	image, err := b.getImage(ref)
	if err != nil {
		return err
	}

	if err := image.Tag(newRef); err != nil {
		return fmt.Errorf("error tagging image: %s", err)
	}

	return nil
}

func (b *NativeRootlessBuildah) Push(ctx context.Context, ref string, opts PushOpts) error {
	pushOpts := buildah.PushOptions{
		Compression:         define.Gzip,
		Store:               b.Store,
		ManifestType:        manifest.DockerV2Schema2MediaType,
		MaxRetries:          MaxPullPushRetries,
		RetryDelay:          PullPushRetryDelay,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext: &imgtypes.SystemContext{
			OCIInsecureSkipTLSVerify:          b.Insecure,
			DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
			DockerDaemonInsecureSkipTLSVerify: b.Insecure,
		},
	}

	if opts.LogWriter != nil {
		pushOpts.ReportWriter = opts.LogWriter
	}

	imageRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", ref))
	if err != nil {
		return fmt.Errorf("error parsing image ref from %q: %s", ref, err)
	}

	if _, _, err = buildah.Push(ctx, ref, imageRef, pushOpts); err != nil {
		return fmt.Errorf("error pushing image %q: %s", ref, err)
	}

	return nil
}

func (b *NativeRootlessBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	buildOpts := define.BuildOptions{
		Isolation:    define.Isolation(b.Isolation),
		OutputFormat: buildah.Dockerv2ImageManifest,
		CommonBuildOpts: &define.CommonBuildOptions{
			ShmSize: DefaultShmSize,
		},
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext: &imgtypes.SystemContext{
			OCIInsecureSkipTLSVerify:          b.Insecure,
			DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
			DockerDaemonInsecureSkipTLSVerify: b.Insecure,
		},
		Args: opts.BuildArgs,
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
		return "", fmt.Errorf("error preparing for build from dockerfile: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(sessionTmpDir); err != nil {
			logboek.Warn().LogF("unable to remove session tmp dir %q: %s\n", sessionTmpDir, err)
		}
	}()
	buildOpts.ContextDirectory = contextTmpDir

	imageId, _, err := imagebuildah.BuildDockerfiles(ctx, b.Store, buildOpts, dockerfileTmpPath)
	if err != nil {
		return "", fmt.Errorf("unable to build Dockerfile %q:\n%s\n%s", dockerfileTmpPath, errLog.String(), err)
	}

	return imageId, nil
}

func (b *NativeRootlessBuildah) Mount(ctx context.Context, container string, opts MountOpts) (string, error) {
	builder, err := b.openContainerBuilder(ctx, container)
	if err != nil {
		return "", fmt.Errorf("unable to open container %q builder: %s", container, err)
	}

	return builder.Mount("")
}

func (b *NativeRootlessBuildah) Umount(ctx context.Context, container string, opts UmountOpts) error {
	builder, err := b.openContainerBuilder(ctx, container)
	if err != nil {
		return fmt.Errorf("unable to open container %q builder: %s", container, err)
	}

	return builder.Unmount()
}

func (b *NativeRootlessBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
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
		return fmt.Errorf("unable to open container %q builder: %s", container, err)
	}

	if err := builder.Run(command, runOpts); err != nil {
		return fmt.Errorf("RunCommand failed:\n%s\n%s", stderr.String(), err)
	}

	return nil
}

func (b *NativeRootlessBuildah) FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) (string, error) {
	builder, err := buildah.NewBuilder(ctx, b.Store, buildah.BuilderOptions{
		FromImage: image,
		Container: container,
	})
	if err != nil {
		return "", fmt.Errorf("unable to create builder: %s", err)
	}

	return builder.Container, builder.Save()
}

func (b *NativeRootlessBuildah) Pull(ctx context.Context, ref string, opts PullOpts) error {
	pullOpts := buildah.PullOptions{
		Store:               b.Store,
		MaxRetries:          MaxPullPushRetries,
		RetryDelay:          PullPushRetryDelay,
		PullPolicy:          define.PullIfNewer,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext: &imgtypes.SystemContext{
			OCIInsecureSkipTLSVerify:          b.Insecure,
			DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
			DockerDaemonInsecureSkipTLSVerify: b.Insecure,
		},
	}

	if opts.LogWriter != nil {
		pullOpts.ReportWriter = opts.LogWriter
	}

	if _, err := buildah.Pull(ctx, ref, pullOpts); err != nil {
		return fmt.Errorf("error pulling image %q: %s", ref, err)
	}

	return nil
}

func (b *NativeRootlessBuildah) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	_, rmiErrors := b.Runtime.RemoveImages(ctx, []string{ref}, &libimage.RemoveImagesOptions{
		Filters: []string{"readonly=false", "intermediate=false", "dangling=true"},
		Force:   opts.Force,
	})

	var multiErr *multierror.Error
	return multierror.Append(multiErr, rmiErrors...).ErrorOrNil()
}

func (b *NativeRootlessBuildah) getImage(ref string) (*libimage.Image, error) {
	image, _, err := b.Runtime.LookupImage(ref, &libimage.LookupImageOptions{
		ManifestList: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error looking up image %q: %s", ref, err)
	}

	return image, nil
}

// getImageBuilder returns nil, nil if image not found.
func (b *NativeRootlessBuildah) getImageBuilder(ctx context.Context, imgName string) (builder *buildah.Builder, err error) {
	builder, err = buildah.ImportBuilderFromImage(ctx, b.Store, buildah.ImportFromImageOptions{
		Image:               imgName,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext: &imgtypes.SystemContext{
			OCIInsecureSkipTLSVerify:          b.Insecure,
			DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
			DockerDaemonInsecureSkipTLSVerify: b.Insecure,
		},
	})
	switch {
	case err != nil && strings.HasSuffix(err.Error(), storage.ErrImageUnknown.Error()):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("error getting builder from image %q: %s", imgName, err)
	case builder == nil:
		panic("error mocking up build configuration")
	}

	return builder, nil
}

func (b *NativeRootlessBuildah) openContainerBuilder(ctx context.Context, container string) (*buildah.Builder, error) {
	builder, err := buildah.OpenBuilder(b.Store, container)
	switch {
	case os.IsNotExist(errors.Cause(err)):
		builder, err = buildah.ImportBuilder(ctx, b.Store, buildah.ImportOptions{
			Container: container,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to import builder for container %q: %s", container, err)
		}
	case err != nil:
		return nil, fmt.Errorf("unable to open builder for container %q: %s", container, err)
	}

	return builder, err
}

func NewNativeStoreOptions(rootlessUID int, driver StorageDriver) (*types.StoreOptions, error) {
	var (
		runRoot string
		err     error
	)

	if rootlessUID == 0 {
		runRoot = "/run/containers/storage"
	} else {
		runRoot, err = storage.GetRootlessRuntimeDir(rootlessUID)
		if err != nil {
			return nil, fmt.Errorf("unable to get runtime dir: %s", err)
		}
	}

	home, err := homedir.GetDataHome()
	if err != nil {
		return nil, fmt.Errorf("unable to get HOME data dir: %s", err)
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
		overlayOpts, err := GetOverlayOptions()
		if err != nil {
			return nil, fmt.Errorf("unable to get overlay options: %s", err)
		}
		graphDriverOptions = append(graphDriverOptions, overlayOpts...)
	}

	return &types.StoreOptions{
		RunRoot:             runRoot,
		GraphRoot:           graphRoot,
		RootlessStoragePath: rootlessStoragePath,
		GraphDriverName:     string(driver),
		GraphDriverOptions:  graphDriverOptions,
	}, nil
}
