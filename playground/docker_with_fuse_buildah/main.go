package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
)

func do(ctx context.Context) error {
	if err := werf.Init("", ""); err != nil {
		return err
	}

	if err := docker.Init(ctx, "", false, false, ""); err != nil {
		return err
	}

	b, err := buildah.NewBuildah(buildah.ModeDockerWithFuse)
	if err != nil {
		return err
	}

	contextFile := filepath.Join(os.Getenv("HOME"), ".go", "src", "github.com", "werf", "werf", "playground", "docker_with_fuse_buildah", "app.tar")
	f, err := os.OpenFile(contextFile, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	imageID, err := b.BuildFromDockerfile(ctx, []byte(`FROM alpine

RUN echo HELLO > /FILE
ADD . /app
`), buildah.BuildFromDockerfileOpts{
		CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout},
		ContextTar: f,
	})
	if err != nil {
		return err
	}

	fmt.Printf("BUILT NEW IMAGE %q\n", imageID)

	if err := b.FromCommand(ctx, "build-container", imageID, buildah.FromCommandOpts{CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout}}); err != nil {
		return err
	}

	if err := b.RunCommand(ctx, "build-container", []string{"ls"}, buildah.RunCommandOpts{CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout}}); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := do(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	}
}
