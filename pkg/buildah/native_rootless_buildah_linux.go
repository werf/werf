//go:build linux
// +build linux

package buildah

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/common/libimage"
	"github.com/containers/image/v5/manifest"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
	"github.com/docker/docker/pkg/reexec"
	"github.com/sirupsen/logrus"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah/types"
	"gopkg.in/errgo.v2/fmt/errors"
)

func InitNativeRootlessProcess() (bool, error) {
	if reexec.Init() {
		return true, nil
	}

	logrus.SetLevel(logrus.TraceLevel)

	unshare.MaybeReexecUsingUserNamespace(false)

	return false, nil
}

type NativeRootlessBuildah struct {
	BaseBuildah

	Store   storage.Store
	Runtime libimage.Runtime
}

func NewNativeRootlessBuildah() (*NativeRootlessBuildah, error) {
	b := &NativeRootlessBuildah{}

	baseBuildah, err := NewBaseBuildah()
	if err != nil {
		return nil, fmt.Errorf("unable to create BaseBuildah: %s", err)
	}
	b.BaseBuildah = *baseBuildah

	storeOpts, err := storage.DefaultStoreOptions(unshare.IsRootless(), unshare.GetRootlessUID())
	if err != nil {
		return nil, fmt.Errorf("unable to set default storage opts: %s", err)
	}
	b.Store, err = storage.GetStore(storeOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to get storage: %s", err)
	}
	is.Transport.SetStore(b.Store)

	runtime, err := libimage.RuntimeFromStore(b.Store, &libimage.RuntimeOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting runtime from store: %s", err)
	}
	b.Runtime = *runtime

	return b, nil
}

func (b *NativeRootlessBuildah) Inspect(ctx context.Context, ref string) (types.BuilderInfo, error) {
	panic("not implemented yet")
}

func (b *NativeRootlessBuildah) Tag(ctx context.Context, ref, newRef string) error {
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
		Compression:  define.Gzip, // REVIEW(ilya-lesikov): compress?
		Store:        b.Store,
		ManifestType: manifest.DockerV2Schema2MediaType, // REVIEW(ilya-lesikov): which one? There was a choice initially:  oci, v2s1, v2s2(Docker).
		MaxRetries:   3,                                 // REVIEW(ilya-lesikov): defaults from buildah
		RetryDelay:   2 * time.Second,                   // REVIEW(ilya-lesikov): defaults from buildah
	}

	if opts.LogWriter != nil {
		pushOpts.ReportWriter = opts.LogWriter
	}

	imageRef, err := alltransports.ParseImageName(ref)
	if err != nil {
		return fmt.Errorf("error parsing image ref from %q: %s", ref, err)
	}

	_, _, err = buildah.Push(ctx, ref, imageRef, pushOpts)
	if err != nil {
		return fmt.Errorf("error pushing image %q: %s", ref, err)
	}

	return nil
}

func (b *NativeRootlessBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	buildOpts := define.BuildOptions{
		Isolation: define.IsolationOCIRootless,
		CommonBuildOpts: &define.CommonBuildOptions{
			ShmSize: DefaultShmSize,
		},
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

func (b *NativeRootlessBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	runOpts := buildah.RunOptions{
		Args: opts.BuildArgs,
	}

	stderr := &bytes.Buffer{}
	if opts.LogWriter != nil {
		runOpts.Stdout = opts.LogWriter
		runOpts.Stderr = io.MultiWriter(opts.LogWriter, stderr)
	} else {
		runOpts.Stderr = stderr
	}

	builder, err := buildah.OpenBuilder(b.Store, container)
	switch {
	case os.IsNotExist(errors.Cause(err)):
		builder, err = buildah.ImportBuilder(ctx, b.Store, buildah.ImportOptions{
			Container: container,
		})
		if err != nil {
			return fmt.Errorf("unable to import builder for container %q: %s", container, err)
		}
	case err != nil:
		return fmt.Errorf("unable to open builder for container %q: %s", container, err)
	}

	if err := builder.Run(command, runOpts); err != nil {
		return fmt.Errorf("RunCommand failed:\n%s\n%s", stderr.String(), err)
	}

	return nil
}

func (b *NativeRootlessBuildah) FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) error {
	panic("not implemented yet")
}

func (b *NativeRootlessBuildah) Pull(ctx context.Context, ref string, opts PullOpts) error {
	panic("not implemented yet")
}

func (b *NativeRootlessBuildah) getImage(ref string) (*libimage.Image, error) {
	image, _, err := b.Runtime.LookupImage(ref, &libimage.LookupImageOptions{
		ManifestList: true,
	})

	return image, fmt.Errorf("error looking up image: %s", err)
}
