// +build linux

package buildah

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
	"github.com/werf/logboek"
	"gopkg.in/errgo.v2/fmt/errors"
)

type NativeRootlessBuildah struct {
	BaseBuildah

	Store storage.Store
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

	return b, nil
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
