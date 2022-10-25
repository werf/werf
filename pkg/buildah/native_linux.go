//go:build linux
// +build linux

package buildah

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

const (
	MaxPullPushRetries = 3
	PullPushRetryDelay = 2 * time.Second
)

var DefaultShell = []string{"/bin/sh", "-c"}

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
	Isolation               thirdparty.Isolation
	TmpDir                  string
	InstanceTmpDir          string
	ConfigTmpDir            string
	SignaturePolicyPath     string
	RegistriesConfigPath    string
	RegistriesConfigDirPath string
	Insecure                bool

	Store                     storage.Store
	Runtime                   libimage.Runtime
	DefaultSystemContext      imgtypes.SystemContext
	DefaultCommonBuildOptions define.CommonBuildOptions

	platforms []struct{ OS, Arch, Variant string }
}

func NewNativeBuildah(commonOpts CommonBuildahOpts, opts NativeModeOpts) (*NativeBuildah, error) {
	b := &NativeBuildah{
		Isolation: *commonOpts.Isolation,
		TmpDir:    commonOpts.TmpDir,
		Insecure:  commonOpts.Insecure,
	}

	if err := os.MkdirAll(b.TmpDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", b.TmpDir, err)
	}

	var err error
	b.InstanceTmpDir, err = ioutil.TempDir(b.TmpDir, "instance")
	if err != nil {
		return nil, fmt.Errorf("unable to create instance tmp dir: %w", err)
	}

	b.ConfigTmpDir = filepath.Join(b.InstanceTmpDir, "config")
	if err := os.MkdirAll(b.ConfigTmpDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", b.ConfigTmpDir, err)
	}

	b.SignaturePolicyPath = filepath.Join(b.ConfigTmpDir, "policy.json")
	if err := ioutil.WriteFile(b.SignaturePolicyPath, []byte(DefaultSignaturePolicy), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write file %q: %w", b.SignaturePolicyPath, err)
	}

	b.RegistriesConfigPath = filepath.Join(b.ConfigTmpDir, "registries.conf")
	if err := ioutil.WriteFile(b.RegistriesConfigPath, []byte(DefaultRegistriesConfig), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write file %q: %w", b.RegistriesConfigPath, err)
	}

	b.RegistriesConfigDirPath = filepath.Join(b.ConfigTmpDir, "registries.conf.d")
	if err := os.MkdirAll(b.RegistriesConfigDirPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", b.RegistriesConfigDirPath, err)
	}

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

	var ulimit []string
	if ulmt := os.Getenv("WERF_BUILDAH_ULIMIT"); ulmt != "" {
		ulimit = strings.Split(ulmt, ",")
	} else {
		rlimits, err := currentRlimits()
		if err != nil {
			return nil, fmt.Errorf("error getting current rlimits: %w", err)
		}
		ulimit = rlimitsToBuildahUlimits(rlimits)
	}

	b.DefaultCommonBuildOptions = define.CommonBuildOptions{
		ShmSize: DefaultShmSize,
		Ulimit:  ulimit,
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

func (b *NativeBuildah) BuildFromDockerfile(ctx context.Context, dockerfile string, opts BuildFromDockerfileOpts) (string, error) {
	buildOpts := define.BuildOptions{
		Isolation:               define.Isolation(b.Isolation),
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

	buildOpts.ContextDirectory = opts.ContextDir

	imageId, _, err := imagebuildah.BuildDockerfiles(ctx, b.Store, buildOpts, dockerfile)
	if err != nil {
		return "", fmt.Errorf("unable to build Dockerfile %q:\n%s\n%w", dockerfile, errLog.String(), err)
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
	builder, err := b.openContainerBuilder(ctx, container)
	if err != nil {
		return fmt.Errorf("unable to open container %q builder: %w", container, err)
	}

	nsOpts, netPolicy := generateNamespaceOptionsAndNetworkPolicy(opts.NetworkType)

	var stdout, stderr io.Writer
	stderrBuf := &bytes.Buffer{}
	if opts.LogWriter != nil {
		stdout = opts.LogWriter
		stderr = io.MultiWriter(opts.LogWriter, stderrBuf)
	} else {
		stderr = stderrBuf
	}

	if opts.PrependShell {
		if len(opts.Shell) > 0 {
			command = append(opts.Shell, command...)
		} else if len(builder.Shell()) > 0 {
			command = append(builder.Shell(), command...)
		} else {
			command = append(DefaultShell, command...)
		}
	}

	runOpts := buildah.RunOptions{
		Env:              opts.Envs,
		ContextDir:       opts.ContextDir,
		AddCapabilities:  opts.AddCapabilities,
		DropCapabilities: opts.DropCapabilities,
		Stdout:           stdout,
		Stderr:           stderr,
		NamespaceOptions: nsOpts,
		ConfigureNetwork: netPolicy,
		Isolation:        define.Isolation(b.Isolation),
		SystemContext:    &b.DefaultSystemContext,
		WorkingDir:       opts.WorkingDir,
		User:             opts.User,
		Entrypoint:       []string{},
		Cmd:              []string{},

		// TODO(ilya-lesikov): implement mounts
		Secrets:             nil,
		SSHSources:          nil,
		Mounts:              opts.LegacyMounts,
		RunMounts:           nil,
		StageMountPoints:    nil,
		ExternalImageMounts: nil,
	}

	if err := builder.Run(command, runOpts); err != nil {
		return fmt.Errorf("RunCommand failed:\n%s\n%w", stderrBuf.String(), err)
	}

	return nil
}

func (b *NativeBuildah) FromCommand(ctx context.Context, container, image string, opts FromCommandOpts) (string, error) {
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

	if opts.Maintainer != "" {
		builder.SetMaintainer(opts.Maintainer)
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

	if len(opts.Shell) > 0 {
		builder.SetShell(opts.Shell)
	}

	if len(opts.Cmd) > 0 {
		var cmd []string
		if opts.CmdPrependShell {
			if builder.Shell() != nil {
				cmd = builder.Shell()
			} else {
				cmd = DefaultShell
			}

			cmd = append(cmd, opts.Cmd...)
		} else {
			cmd = opts.Cmd
		}

		builder.SetCmd(cmd)
	}

	if len(opts.Entrypoint) > 0 {
		var entrypoint []string
		if opts.EntrypointPrependShell {
			if builder.Shell() != nil {
				entrypoint = builder.Shell()
			} else {
				entrypoint = DefaultShell
			}

			entrypoint = append(entrypoint, opts.Entrypoint...)
		} else {
			entrypoint = opts.Entrypoint
		}

		builder.SetEntrypoint(entrypoint)
	}

	if opts.User != "" {
		builder.SetUser(opts.User)
	}

	if opts.Workdir != "" {
		builder.SetWorkDir(opts.Workdir)
	}

	if opts.Healthcheck != nil {
		builder.SetHealthcheck((*docker.HealthConfig)(opts.Healthcheck))
	}

	if opts.StopSignal != "" {
		builder.SetStopSignal(opts.StopSignal)
	}

	if opts.OnBuild != "" {
		builder.SetOnBuild(opts.OnBuild)
	}

	return builder.Save()
}

func (b *NativeBuildah) Copy(ctx context.Context, container, contextDir string, src []string, dst string, opts CopyOpts) error {
	builder, err := b.getBuilderFromContainer(ctx, container)
	if err != nil {
		return fmt.Errorf("error getting builder: %w", err)
	}

	var absSrc []string
	for _, s := range src {
		absSrc = append(absSrc, filepath.Join(contextDir, s))
	}

	if err := builder.Add(dst, false, buildah.AddAndCopyOptions{
		Chown:             opts.Chown,
		Chmod:             opts.Chmod,
		PreserveOwnership: false,
		ContextDir:        contextDir,
		// TODO(ilya-lesikov): ignore file?
		Excludes: nil,
		// TODO(ilya-lesikov): ignore file?
		IgnoreFile: "",
	}, absSrc...); err != nil {
		return fmt.Errorf("error copying files to %q: %w", dst, err)
	}

	return nil
}

func (b *NativeBuildah) Add(ctx context.Context, container string, src []string, dst string, opts AddOpts) error {
	builder, err := b.getBuilderFromContainer(ctx, container)
	if err != nil {
		return fmt.Errorf("error getting builder: %w", err)
	}

	var expandedSrc []string
	for _, s := range src {
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			expandedSrc = append(expandedSrc, s)
			continue
		}

		if opts.ContextDir == "" {
			return fmt.Errorf("context dir is required for adding local files")
		}

		expandedSrc = append(expandedSrc, filepath.Join(opts.ContextDir, s))
	}

	if err := builder.Add(dst, true, buildah.AddAndCopyOptions{
		Chmod:             opts.Chmod,
		Chown:             opts.Chown,
		PreserveOwnership: false,
		ContextDir:        opts.ContextDir,
		// TODO(ilya-lesikov): ignore file?
		Excludes: nil,
		// TODO(ilya-lesikov): ignore file?
		IgnoreFile: "",
	}, expandedSrc...); err != nil {
		return fmt.Errorf("error adding files to %q: %w", dst, err)
	}

	return nil
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

func (b *NativeBuildah) NewSessionTmpDir() (string, error) {
	sessionTmpDir, err := ioutil.TempDir(b.TmpDir, "session")
	if err != nil {
		return "", fmt.Errorf("unable to create session tmp dir: %w", err)
	}

	return sessionTmpDir, nil
}

func (b *NativeBuildah) prepareBuildFromDockerfile(dockerfile []byte, contextTar io.Reader) (string, string, string, error) {
	sessionTmpDir, err := b.NewSessionTmpDir()
	if err != nil {
		return "", "", "", err
	}

	dockerfileTmpPath := filepath.Join(sessionTmpDir, "Dockerfile")
	if err := ioutil.WriteFile(dockerfileTmpPath, dockerfile, os.ModePerm); err != nil {
		return "", "", "", fmt.Errorf("error writing %q: %w", dockerfileTmpPath, err)
	}

	contextTmpDir := filepath.Join(sessionTmpDir, "context")
	if err := os.MkdirAll(contextTmpDir, os.ModePerm); err != nil {
		return "", "", "", fmt.Errorf("unable to create dir %q: %w", contextTmpDir, err)
	}

	if contextTar != nil {
		if err := util.ExtractTar(contextTar, contextTmpDir, util.ExtractTarOptions{}); err != nil {
			return "", "", "", fmt.Errorf("unable to extract context tar to tmp context dir: %w", err)
		}
	}

	return sessionTmpDir, contextTmpDir, dockerfileTmpPath, nil
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

func currentRlimits() (map[int]*syscall.Rlimit, error) {
	result := map[int]*syscall.Rlimit{
		syscall.RLIMIT_CORE:   {},
		syscall.RLIMIT_CPU:    {},
		syscall.RLIMIT_DATA:   {},
		syscall.RLIMIT_FSIZE:  {},
		syscall.RLIMIT_NOFILE: {},
		syscall.RLIMIT_STACK:  {},
	}

	for k, v := range result {
		if err := syscall.Getrlimit(k, v); err != nil {
			return nil, fmt.Errorf("error getting rlimit: %w", err)
		}
	}

	return result, nil
}

func rlimitsToBuildahUlimits(rlimits map[int]*syscall.Rlimit) []string {
	rlimitToBuildahUlimitFn := func(rlimitKey int, buildahKey string) string {
		rlimitUintToStrFn := func(val uint64) string {
			if int64(val) < 0 {
				return strconv.FormatInt(int64(val), 10)
			} else {
				return strconv.FormatUint(val, 10)
			}
		}

		cur := rlimitUintToStrFn(rlimits[rlimitKey].Cur)
		max := rlimitUintToStrFn(rlimits[rlimitKey].Max)
		return fmt.Sprintf("%s=%s:%s", buildahKey, cur, max)
	}

	return []string{
		rlimitToBuildahUlimitFn(syscall.RLIMIT_CORE, "core"),
		rlimitToBuildahUlimitFn(syscall.RLIMIT_CPU, "cpu"),
		rlimitToBuildahUlimitFn(syscall.RLIMIT_DATA, "data"),
		rlimitToBuildahUlimitFn(syscall.RLIMIT_FSIZE, "fsize"),
		rlimitToBuildahUlimitFn(syscall.RLIMIT_NOFILE, "nofile"),
		rlimitToBuildahUlimitFn(syscall.RLIMIT_STACK, "stack"),
	}
}

func generateNamespaceOptionsAndNetworkPolicy(network instruction.NetworkType) (define.NamespaceOptions, define.NetworkConfigurationPolicy) {
	var netPolicy define.NetworkConfigurationPolicy
	nsOpts := define.NamespaceOptions{}

	switch network {
	case instruction.NetworkDefault, "":
		netPolicy = define.NetworkDefault
		nsOpts.AddOrReplace(define.NamespaceOption{
			Name: string(specs.NetworkNamespace),
		})
	case instruction.NetworkHost:
		netPolicy = define.NetworkEnabled
		nsOpts.AddOrReplace(define.NamespaceOption{
			Name: string(specs.NetworkNamespace),
			Host: true,
		})
	case instruction.NetworkNone:
		netPolicy = define.NetworkDisabled
		nsOpts.AddOrReplace(define.NamespaceOption{
			Name: string(specs.NetworkNamespace),
		})
	default:
		panic(fmt.Sprintf("unexpected network type: %v", network))
	}

	return nsOpts, netPolicy
}

// TODO(ilya-lesikov): implement mounts
// func generateMounts(mounts []*instructions.Mount, contextDir string) (map[string]*specs.Mount, error) {
// 	result := map[string]*specs.Mount{}
//
// 	for _, mount := range mounts {
// 		switch mount.Type {
// 		case instructions.MountTypeBind:
// 			if m, err := generateBindMount(mount, contextDir); err != nil {
// 				return nil, fmt.Errorf("error generating bind mount: %w", err)
// 			} else {
// 				result[mount.Target] = m
// 			}
// 		case instructions.MountTypeCache:
// 			result[mount.Target] = &specs.Mount{
// 				Type:        parse.TypeCache,
// 				Source:      mount.Source,
// 				Destination: mount.Target,
// 				Options:     generateMountOptions(mount),
// 			}
// 		case instructions.MountTypeTmpfs:
// 		case instructions.MountTypeSecret:
// 		case instructions.MountTypeSSH:
// 		}
// 	}
// }
//
// func generateBindMount(mount *instructions.Mount, contextDir string) (*specs.Mount, string, error) {
// 	var mountedImage string
// 	var source string
// 	if mount.From != "" && contextDir != "" {
// 		// FIXME(ilya-lesikov): here also be container mounts and new contextDir:
// 		// /home/user1/go/pkg/mod/github.com/containers/buildah@v1.26.1/internal/parse/parse.go:117
// 		source = filepath.Join(contextDir, filepath.Clean(string(filepath.Separator)+mount.Source))
// 		mountedImage = "" // TODO
// 	} else {
// 		if err := parse.ValidateVolumeHostDir(mount.Source); err != nil {
// 			return nil, "", fmt.Errorf("invalid bind mount source %q: %w", mount.Source, err)
// 		}
// 		source = mount.Source
// 	}
//
// 	if err := parse.ValidateVolumeCtrDir(mount.Target); err != nil {
// 		return nil, "", fmt.Errorf("invalid bind mount target %q: %w", mount.Target, err)
// 	}
//
// 	options := []string{"rbind"}
//
// 	if mount.ReadOnly {
// 		options = append(options, "ro")
// 	} else {
// 		options = append(options, "rw")
// 	}
//
// 	var err error
// 	options, err = parse.ValidateVolumeOpts(options)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("invalid bind mount options %v: %w", options, err)
// 	}
//
// 	return &specs.Mount{
// 		Type:        parse.TypeBind,
// 		Source:      source,
// 		Destination: mount.Target,
// 		Options:     options,
// 	}, mountedImage, nil
// }
//
// func generateCacheMount(mount *instructions.Mount) (*specs.Mount, error) {
// 	if err := parse.ValidateVolumeCtrDir(mount.Target); err != nil {
// 		return nil, fmt.Errorf("invalid cache mount target %q: %w", mount.Target, err)
// 	}
//
// 	options := []string{"bind", "shared"}
//
// 	if mount.ReadOnly {
// 		options = append(options, "ro")
// 	} else {
// 		options = append(options, "rw")
// 	}
//
// 	// var mode uint64
// 	// if mount.Mode != nil {
// 	// 	mode = *mount.Mode
// 	// } else {
// 	// 	mode = 0o755
// 	// }
// 	//
// 	// var uid uint64
// 	// if mount.UID != nil {
// 	// 	uid = *mount.UID
// 	// } else {
// 	// 	uid = 0
// 	// }
// 	//
// 	// var gid uint64
// 	// if mount.GID != nil {
// 	// 	gid = *mount.GID
// 	// } else {
// 	// 	gid = 0
// 	// }
//
// 	// TODO(ilya-lesikov): create cache dir/image/stage, example:
// 	//  /home/user1/go/pkg/mod/github.com/containers/buildah@v1.26.1/internal/parse/parse.go:286
//
// 	if mount.CacheSharing == "locked" {
// 		// TODO(ilya-lesikov): Implement cache access locking
// 		//  /home/user1/go/pkg/mod/github.com/containers/buildah@v1.26.1/internal/parse/parse.go:335
// 	}
//
// 	var err error
// 	options, err = parse.ValidateVolumeOpts(options)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid cache mount options %v: %w", options, err)
// 	}
//
// 	return &specs.Mount{
// 		Type:        parse.TypeBind,
// 		Source:      mount.Source,
// 		Destination: mount.Target,
// 		Options:     options,
// 	}, nil
// }
//
// func generateTmpFsMount(mount *instructions.Mount) (*specs.Mount, error) {
// 	if err := parse.ValidateVolumeCtrDir(mount.Target); err != nil {
// 		return nil, fmt.Errorf("invalid tmpfs mount target %q: %w", mount.Target, err)
// 	}
//
// 	options := []string{"bind", "shared"}
//
// 	if mount.ReadOnly {
// 		options = append(options, "ro")
// 	} else {
// 		options = append(options, "rw")
// 	}
//
// 	if mount.Mode != nil {
// 		options = append(options, fmt.Sprintf("mode=%o", *mount.Mode))
// 	}
//
// 	var err error
// 	options, err = parse.ValidateVolumeOpts(options)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid tmpfs mount options %v: %w", options, err)
// 	}
//
// 	return &specs.Mount{
// 		Type:        parse.TypeTmpfs,
// 		Source:      parse.TypeTmpfs,
// 		Destination: mount.Target,
// 		Options:     options,
// 	}, nil
// }
//
// func getSecretMount(mount *instructions.Mount) (*specs.Mount, error) {
// 	if err := parse.ValidateVolumeCtrDir(mount.Target); err != nil {
// 		return nil, fmt.Errorf("invalid secret mount target %q: %w", mount.Target, err)
// 	}
//
// 	options := []string{"bind", "shared"}
//
// 	if mount.ReadOnly {
// 		options = append(options, "ro")
// 	} else {
// 		options = append(options, "rw")
// 	}
//
// 	var err error
// 	options, err = parse.ValidateVolumeOpts(options)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid secret mount options %v: %w", options, err)
// 	}
//
// 	return &specs.Mount{
// 		Type:        parse.TypeSecret,
// 		Source:      mount.Source,
// 		Destination: mount.Target,
// 		Options:     options,
// 	}, nil
// }
