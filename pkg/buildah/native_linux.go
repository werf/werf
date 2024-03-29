//go:build linux
// +build linux

package buildah

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd/platforms"
	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/docker"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/buildah/pkg/parse"
	"github.com/containers/common/libimage"
	"github.com/containers/image/v5/manifest"
	imgstor "github.com/containers/image/v5/storage"
	storageTransport "github.com/containers/image/v5/storage"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	imgtypes "github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/containers/storage/drivers/overlay"
	"github.com/containers/storage/pkg/homedir"
	"github.com/containers/storage/pkg/reexec"
	"github.com/containers/storage/pkg/unshare"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/image"
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
	ContainersConfigPath    string
	RegistriesConfigPath    string
	RegistriesConfigDirPath string
	Insecure                bool

	Store storage.Store

	defaultCommonBuildOptions    define.CommonBuildOptions
	defaultSystemContext         imgtypes.SystemContext
	defaultPlatform              string
	defaultPlatformOverrideSpecs []struct{ OS, Arch, Variant string }
	runtimePlatform              string
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

	b.ContainersConfigPath = filepath.Join(b.ConfigTmpDir, "containers.conf")
	if err := ioutil.WriteFile(b.ContainersConfigPath, []byte(DefaultContainersConfig), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write file %q: %w", b.ContainersConfigPath, err)
	}

	if err := os.Setenv("CONTAINERS_CONF", b.ContainersConfigPath); err != nil {
		return nil, fmt.Errorf("unable to set env var CONTAINERS_CONF: %w", err)
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

	b.defaultSystemContext = imgtypes.SystemContext{
		SignaturePolicyPath:               b.SignaturePolicyPath,
		SystemRegistriesConfPath:          b.RegistriesConfigPath,
		SystemRegistriesConfDirPath:       b.RegistriesConfigDirPath,
		OCIInsecureSkipTLSVerify:          b.Insecure,
		DockerInsecureSkipTLSVerify:       imgtypes.NewOptionalBool(b.Insecure),
		DockerDaemonInsecureSkipTLSVerify: b.Insecure,
	}

	b.runtimePlatform = platforms.Format(platforms.DefaultSpec())

	if opts.DefaultPlatform != "" {
		b.defaultPlatform = opts.DefaultPlatform

		os, arch, variant, err := parse.Platform(opts.DefaultPlatform)
		if err != nil {
			return nil, fmt.Errorf("unable to parse platform %q: %w", opts.DefaultPlatform, err)
		}

		b.defaultSystemContext.OSChoice = os
		b.defaultSystemContext.ArchitectureChoice = arch
		b.defaultSystemContext.VariantChoice = variant

		b.defaultPlatformOverrideSpecs = []struct{ OS, Arch, Variant string }{
			{os, arch, variant},
		}
	} else {
		b.defaultPlatform = b.runtimePlatform
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

	b.defaultCommonBuildOptions = define.CommonBuildOptions{
		ShmSize: DefaultShmSize,
		Ulimit:  ulimit,
	}

	imgstor.Transport.SetStore(b.Store)

	return b, nil
}

func (b *NativeBuildah) getSystemContext(targetPlatform string) (*imgtypes.SystemContext, error) {
	systemContext := new(imgtypes.SystemContext)
	*systemContext = b.defaultSystemContext

	if targetPlatform != "" {
		os, arch, variant, err := parse.Platform(targetPlatform)
		if err != nil {
			return nil, fmt.Errorf("unable to parse platform %q: %w", targetPlatform, err)
		}

		systemContext.OSChoice = os
		systemContext.ArchitectureChoice = arch
		systemContext.VariantChoice = variant
	}

	return systemContext, nil
}

func (b *NativeBuildah) getRuntime(systemContext *imgtypes.SystemContext) (*libimage.Runtime, error) {
	return libimage.RuntimeFromStore(b.Store, &libimage.RuntimeOptions{
		SystemContext: systemContext,
	})
}

func (b *NativeBuildah) GetRuntimePlatform() string {
	return b.runtimePlatform
}

func (b *NativeBuildah) GetDefaultPlatform() string {
	return b.defaultPlatform
}

// Inspect returns nil, nil if image not found.
func (b *NativeBuildah) Inspect(ctx context.Context, ref string) (*thirdparty.BuilderInfo, error) {
	builder, err := b.getBuilderFromImage(ctx, ref, CommonOpts{})
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
	image, err := b.getImage(ref, CommonOpts(opts))
	if err != nil {
		return err
	}

	if err := image.Tag(newRef); err != nil {
		return fmt.Errorf("error tagging image: %w", err)
	}

	return nil
}

func (b *NativeBuildah) Push(ctx context.Context, ref string, opts PushOpts) error {
	// NOTICE: targetPlatform specified for push causes buildah to fail for some unknown reason
	sysCtx, err := b.getSystemContext("")
	if err != nil {
		return err
	}

	pushOpts := buildah.PushOptions{
		Compression:         define.Gzip,
		SignaturePolicyPath: b.SignaturePolicyPath,
		ReportWriter:        opts.LogWriter,
		Store:               b.Store,
		SystemContext:       sysCtx,
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
	var targetPlatform string
	var targetPlatforms []struct{ OS, Arch, Variant string }
	if opts.TargetPlatform != "" {
		os, arch, variant, err := parse.Platform(opts.TargetPlatform)
		if err != nil {
			return "", fmt.Errorf("unable to parse target platform %q: %w", opts.TargetPlatform, err)
		}

		targetPlatform = opts.TargetPlatform
		targetPlatforms = []struct{ OS, Arch, Variant string }{
			{os, arch, variant},
		}
	} else {
		targetPlatform = b.defaultPlatform
		targetPlatforms = b.defaultPlatformOverrideSpecs
	}

	// FIXME(multiarch): Fix warning in the build log:
	// FIXME(multiarch):  [Warning] one or more build args were not consumed: [TARGETARCH TARGETOS TARGETPLATFORM]

	sysCtx, err := b.getSystemContext(targetPlatform)
	if err != nil {
		return "", err
	}

	buildOpts := define.BuildOptions{
		Isolation:               define.Isolation(b.Isolation),
		Args:                    opts.BuildArgs,
		SignaturePolicyPath:     b.SignaturePolicyPath,
		ReportWriter:            opts.LogWriter,
		OutputFormat:            buildah.Dockerv2ImageManifest,
		SystemContext:           sysCtx,
		ConfigureNetwork:        define.NetworkEnabled,
		CommonBuildOpts:         &b.defaultCommonBuildOptions,
		Target:                  opts.Target,
		Platforms:               targetPlatforms,
		MaxPullPushRetries:      MaxPullPushRetries,
		PullPushRetryDelay:      PullPushRetryDelay,
		Layers:                  true,
		RemoveIntermediateCtrs:  true,
		ForceRmIntermediateCtrs: false,
		NoCache:                 false,
		Labels:                  opts.Labels,
	}

	if targetPlatform != b.GetRuntimePlatform() {
		// Prevent local cache collisions in multiplatform build mode:
		//   allow local cache only for the current runtime platform.
		buildOpts.NoCache = true
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

	contextDir := generateContextDir(opts.ContextDir, opts.RunMounts)
	nsOpts, netPolicy := generateNamespaceOptionsAndNetworkPolicy(opts.NetworkType)
	globalMounts := generateGlobalMounts(opts.GlobalMounts)
	runMounts := generateRunMounts(opts.RunMounts)
	stdout, stderr, stderrBuf := generateStdoutStderr(opts.LogWriter)
	command = prependShellToCommand(opts.PrependShell, opts.Shell, command, builder)

	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return err
	}

	runOpts := buildah.RunOptions{
		Env:              opts.Envs,
		ContextDir:       contextDir,
		AddCapabilities:  opts.AddCapabilities,
		DropCapabilities: opts.DropCapabilities,
		Stdout:           stdout,
		Stderr:           stderr,
		NamespaceOptions: nsOpts,
		ConfigureNetwork: netPolicy,
		Isolation:        define.Isolation(b.Isolation),
		SystemContext:    sysCtx,
		WorkingDir:       opts.WorkingDir,
		User:             opts.User,
		Entrypoint:       []string{},
		Cmd:              []string{},
		Mounts:           globalMounts,
		RunMounts:        runMounts,
		// TODO(ilya-lesikov):
		Secrets: nil,
		// TODO(ilya-lesikov):
		SSHSources: nil,
	}

	if err := builder.Run(command, runOpts); err != nil {
		return fmt.Errorf("RunCommand failed:\n%s\n%w", stderrBuf.String(), err)
	}

	return nil
}

func (b *NativeBuildah) FromCommand(ctx context.Context, container, image string, opts FromCommandOpts) (string, error) {
	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return "", err
	}

	builder, err := buildah.NewBuilder(ctx, b.Store, buildah.BuilderOptions{
		FromImage:           image,
		Container:           container,
		SignaturePolicyPath: b.SignaturePolicyPath,
		ReportWriter:        opts.LogWriter,
		SystemContext:       sysCtx,
		Isolation:           define.Isolation(b.Isolation),
		ConfigureNetwork:    define.NetworkEnabled,
		CommonBuildOpts:     &b.defaultCommonBuildOptions,
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
	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return err
	}

	pullOpts := buildah.PullOptions{
		SignaturePolicyPath: b.SignaturePolicyPath,
		ReportWriter:        opts.LogWriter,
		Store:               b.Store,
		SystemContext:       sysCtx,
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
	if sysCtx.OSChoice != "" && sysCtx.OSChoice != imageInspect.OCIv1.OS {
		platformMismatch = true
	}
	if sysCtx.ArchitectureChoice != "" && sysCtx.ArchitectureChoice != imageInspect.OCIv1.Architecture {
		platformMismatch = true
	}

	if platformMismatch {
		imagePlatform := fmt.Sprintf("%s/%s", imageInspect.OCIv1.OS, imageInspect.OCIv1.Architecture)
		if imageInspect.OCIv1.Variant != "" {
			imagePlatform = fmt.Sprintf("%s/%s", imagePlatform, imageInspect.OCIv1.Variant)
		}
		expectedPlatform := fmt.Sprintf("%s/%s", sysCtx.OSChoice, sysCtx.ArchitectureChoice)
		if sysCtx.VariantChoice != "" {
			expectedPlatform = fmt.Sprintf("%s/%s", expectedPlatform, sysCtx.VariantChoice)
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
	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return err
	}

	runtime, err := b.getRuntime(sysCtx)
	if err != nil {
		return err
	}

	_, rmiErrors := runtime.RemoveImages(ctx, []string{ref}, &libimage.RemoveImagesOptions{
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

	var imageRef types.ImageReference
	if opts.Image != "" {
		normalizedImage, err := libimage.NormalizeName(opts.Image)
		if err != nil {
			return "", fmt.Errorf("normalizing target image name %q: %w", opts.Image, err)
		}
		imageRef, err = storageTransport.Transport.ParseStoreReference(b.Store, normalizedImage.String())
		if err != nil {
			return "", fmt.Errorf("parsing target image name %q: %w", opts.Image, err)
		}
	}

	builder.SetLabel(image.WerfBaseImageIDLabel, fmt.Sprintf("sha256:%s", builder.FromImageID))

	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return "", err
	}

	imgID, _, _, err := builder.Commit(ctx, imageRef, buildah.CommitOptions{
		PreferredManifestType: buildah.Dockerv2ImageManifest,
		SignaturePolicyPath:   b.SignaturePolicyPath,
		ReportWriter:          opts.LogWriter,
		SystemContext:         sysCtx,
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
		Excludes:          opts.Ignores,
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
		Excludes:          opts.Ignores,
	}, expandedSrc...); err != nil {
		return fmt.Errorf("error adding files to %q: %w", dst, err)
	}

	return nil
}

func (b *NativeBuildah) getImage(ref string, opts CommonOpts) (*libimage.Image, error) {
	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return nil, err
	}

	runtime, err := b.getRuntime(sysCtx)
	if err != nil {
		return nil, err
	}

	image, _, err := runtime.LookupImage(ref, &libimage.LookupImageOptions{
		ManifestList: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error looking up image %q: %w", ref, err)
	}

	return image, nil
}

// getBuilderFromImage returns nil, nil if image not found.
func (b *NativeBuildah) getBuilderFromImage(ctx context.Context, imgName string, opts CommonOpts) (builder *buildah.Builder, err error) {
	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return nil, err
	}

	builder, err = buildah.ImportBuilderFromImage(ctx, b.Store, buildah.ImportFromImageOptions{
		Image:               imgName,
		SignaturePolicyPath: b.SignaturePolicyPath,
		SystemContext:       sysCtx,
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

func (b *NativeBuildah) Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error) {
	sysCtx, err := b.getSystemContext(opts.TargetPlatform)
	if err != nil {
		return nil, err
	}

	runtime, err := b.getRuntime(sysCtx)
	if err != nil {
		return nil, err
	}

	listOpts := &libimage.ListImagesOptions{}
	for _, filter := range opts.Filters {
		listOpts.Filters = append(listOpts.Filters, fmt.Sprintf("%s=%s", filter.First, filter.Second))
	}
	images, err := runtime.ListImages(ctx, opts.Names, listOpts)
	if err != nil {
		return nil, err
	}

	var res image.ImagesList
	for _, img := range images {
		repoTags, err := img.RepoTags()
		if err != nil {
			return nil, fmt.Errorf("unable to get image %s repo tags: %w", img.ID(), err)
		}
		repoDigests, err := img.RepoDigests()
		if err != nil {
			return nil, fmt.Errorf("unable to get image %s repo digests: %w", img.ID(), err)
		}
		res = append(res, image.Summary{RepoTags: repoTags, RepoDigests: repoDigests})
	}

	return res, nil
}

func (b *NativeBuildah) Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error) {
	builders, err := buildah.OpenAllBuilders(b.Store)
	if err != nil {
		return nil, err
	}

	seenImages := make(map[string]string)
	imageNameForID := func(id string) string {
		if id == "" {
			return buildah.BaseImageFakeName
		}
		imageName, ok := seenImages[id]
		if ok {
			return imageName
		}
		img, err2 := b.Store.Image(id)
		if err2 == nil && len(img.Names) > 0 {
			seenImages[id] = img.Names[0]
		}
		return seenImages[id]
	}

	var res image.ContainerList
SelectContainers:
	for _, builder := range builders {
		imgID := builder.FromImageID
		imgName := imageNameForID(builder.FromImageID)

		for _, filter := range opts.Filters {
			if filter.ID != "" && !thirdparty.MatchesID(builder.ContainerID, filter.ID) {
				continue SelectContainers
			}
			if filter.Name != "" && !thirdparty.MatchesCtrName(builder.Container, filter.Name) {
				continue SelectContainers
			}
			if filter.Ancestor != "" && !thirdparty.MatchesAncestor(imgName, imgID, filter.Ancestor) {
				continue SelectContainers
			}
		}

		res = append(res, image.Container{
			ID:      builder.ContainerID,
			ImageID: builder.FromImageID,
			Names:   []string{builder.Container},
		})
	}

	return res, nil
}

func NewNativeStoreOptions(rootlessUID int, driver StorageDriver) (*thirdparty.StoreOptions, error) {
	var (
		runRoot string
		err     error
	)

	if rootlessUID == 0 {
		runRoot = "/run/containers/storage"
	} else {
		runRoot, err = homedir.GetRuntimeDir()
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

func generateNamespaceOptionsAndNetworkPolicy(network string) (define.NamespaceOptions, define.NetworkConfigurationPolicy) {
	var netPolicy define.NetworkConfigurationPolicy
	nsOpts := define.NamespaceOptions{}

	switch network {
	case "default", "":
		netPolicy = define.NetworkDefault
		nsOpts.AddOrReplace(define.NamespaceOption{
			Name: string(specs.NetworkNamespace),
		})
	case "host":
		netPolicy = define.NetworkEnabled
		nsOpts.AddOrReplace(define.NamespaceOption{
			Name: string(specs.NetworkNamespace),
			Host: true,
		})
	case "none":
		netPolicy = define.NetworkDisabled
		nsOpts.AddOrReplace(define.NamespaceOption{
			Name: string(specs.NetworkNamespace),
		})
	default:
		panic(fmt.Sprintf("unexpected network type: %v", network))
	}

	return nsOpts, netPolicy
}

func generateRunMounts(mounts []*instructions.Mount) []string {
	var runMounts []string

	for _, mount := range mounts {
		var options []string

		switch mount.Type {
		case instructions.MountTypeBind:
			options = append(
				options,
				fmt.Sprintf("type=%s", mount.Type),
				fmt.Sprintf("target=%s", mount.Target),
			)
			if mount.Source != "" {
				options = append(options, fmt.Sprintf("source=%s", mount.Source))
			}
			if mount.From != "" {
				options = append(options, fmt.Sprintf("from=%s", mount.From))
			}
			if mount.ReadOnly {
				options = append(options, "ro")
			} else {
				options = append(options, "rw")
			}
		case instructions.MountTypeCache:
			options = append(
				options,
				fmt.Sprintf("type=%s", mount.Type),
				fmt.Sprintf("target=%s", mount.Target),
			)
			if mount.CacheID != "" {
				options = append(options, fmt.Sprintf("id=%s", mount.CacheID))
			}
			if mount.ReadOnly {
				options = append(options, "ro")
			} else {
				options = append(options, "rw")
			}
			if mount.CacheSharing != "" {
				options = append(options, fmt.Sprintf("sharing=%s", mount.CacheSharing))
			}
			if mount.From != "" {
				options = append(options, fmt.Sprintf("from=%s", mount.From))
			}
			if mount.Source != "" {
				options = append(options, fmt.Sprintf("source=%s", mount.Source))
			}
			if mount.Mode != nil {
				options = append(options, fmt.Sprintf("mode=%d", *mount.Mode))
			}
			if mount.UID != nil {
				options = append(options, fmt.Sprintf("uid=%d", *mount.UID))
			}
			if mount.GID != nil {
				options = append(options, fmt.Sprintf("gid=%d", *mount.GID))
			}
		case instructions.MountTypeTmpfs:
			options = append(
				options,
				fmt.Sprintf("type=%s", mount.Type),
				fmt.Sprintf("target=%s", mount.Target),
			)
		case instructions.MountTypeSecret:
			options = append(
				options,
				fmt.Sprintf("type=%s", mount.Type),
			)
			if mount.CacheID != "" {
				options = append(options, fmt.Sprintf("id=%s", mount.CacheID))
			}
			if mount.Target != "" {
				options = append(options, fmt.Sprintf("target=%s", mount.Target))
			}
			if mount.Required {
				options = append(options, "required=true")
			}
			if mount.Mode != nil {
				options = append(options, fmt.Sprintf("mode=%d", *mount.Mode))
			}
			if mount.UID != nil {
				options = append(options, fmt.Sprintf("uid=%d", *mount.UID))
			}
			if mount.GID != nil {
				options = append(options, fmt.Sprintf("gid=%d", *mount.GID))
			}
		case instructions.MountTypeSSH:
			options = append(
				options,
				fmt.Sprintf("type=%s", mount.Type),
			)
			if mount.CacheID != "" {
				options = append(options, fmt.Sprintf("id=%s", mount.CacheID))
			}
			if mount.Target != "" {
				options = append(options, fmt.Sprintf("target=%s", mount.Target))
			}
			if mount.Required {
				options = append(options, "required=true")
			}
			if mount.Mode != nil {
				options = append(options, fmt.Sprintf("mode=%d", *mount.Mode))
			}
			if mount.UID != nil {
				options = append(options, fmt.Sprintf("uid=%d", *mount.UID))
			}
			if mount.GID != nil {
				options = append(options, fmt.Sprintf("gid=%d", *mount.GID))
			}
		default:
			panic(fmt.Sprintf("unexpected mount type %q", mount.Type))
		}

		runMounts = append(runMounts, strings.Join(options, ","))
	}

	return runMounts
}

func generateContextDir(rawContextDir string, runMounts []*instructions.Mount) string {
	usesBuildContext := false
	for _, mount := range runMounts {
		if mount.Type == instructions.MountTypeBind && mount.From == "" {
			usesBuildContext = true
			break
		}
	}

	var contextDir string
	if !usesBuildContext && rawContextDir == "" {
		// Buildah API requires ContextDir even if not actually used. In this case we will pass a dummy value.
		contextDir = strconv.Itoa(rand.Int())
	} else {
		contextDir = rawContextDir
	}

	return contextDir
}

func generateStdoutStderr(optionalLogWriter io.Writer) (stdout, stderr io.Writer, stderrBuf *bytes.Buffer) {
	stderrBuf = &bytes.Buffer{}
	if optionalLogWriter != nil {
		stdout = optionalLogWriter
		stderr = io.MultiWriter(optionalLogWriter, stderrBuf)
	} else {
		stderr = stderrBuf
	}

	return stdout, stderr, stderrBuf
}

func prependShellToCommand(prependShell bool, shell, command []string, builder *buildah.Builder) []string {
	if !prependShell {
		return command
	}

	if len(shell) > 0 {
		command = append(shell, command...)
	} else if len(builder.Shell()) > 0 {
		command = append(builder.Shell(), command...)
	} else {
		command = append(DefaultShell, command...)
	}

	return command
}

func generateGlobalMounts(rawGlobalMounts []*specs.Mount) []specs.Mount {
	var globalMounts []specs.Mount
	for _, mount := range rawGlobalMounts {
		globalMounts = append(globalMounts, *mount)
	}

	return globalMounts
}
