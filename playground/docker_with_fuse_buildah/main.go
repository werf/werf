package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"

	"github.com/werf/werf/pkg/util"

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

	if len(os.Args) != 3 {
		return fmt.Errorf("usage: %s DOCKERFILE_PATH CONTEXT_DIR", os.Args[0])
	}
	dockerfilePath := os.Args[1]
	if dockerfilePath == "" {
		return fmt.Errorf("usage: %s DOCKERFILE_PATH CONTEXT_DIR", os.Args[0])
	}
	contextDir := os.Args[2]
	if contextDir == "" {
		return fmt.Errorf("usage: %s DOCKERFILE_PATH CONTEXT_DIR", os.Args[0])
	}

	dockerfileData, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("error reading %q: %s", dockerfilePath, err)
	}

	contextTar := util.ReadDirAsTar(contextDir)

	imageID, err := b.BuildFromDockerfile(ctx, dockerfileData, buildah.BuildFromDockerfileOpts{
		CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout},
		ContextTar: contextTar,
	})
	if err != nil {
		return err
	}

	fmt.Printf("BUILT NEW IMAGE %q\n", imageID)

	containerName := uuid.New().String()

	if err := b.FromCommand(ctx, containerName, imageID, buildah.FromCommandOpts{CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout}}); err != nil {
		return err
	}

	if err := b.RunCommand(ctx, containerName, []string{"ls"}, buildah.RunCommandOpts{CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout}}); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := do(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
