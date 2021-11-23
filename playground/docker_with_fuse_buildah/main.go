package main

import (
	"context"
	"fmt"
	"io"
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

	b, err := buildah.NewBuildah(buildah.ModeDockerWithFuse, buildah.BuildahOpts{})
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

	errCh := make(chan error, 0)
	buildDoneCh := make(chan string, 0)

	contextTar := util.BufferedPipedWriterProcess(func(w io.WriteCloser) {
		if err := util.WriteDirAsTar((contextDir), w); err != nil {
			errCh <- fmt.Errorf("unable to write dir %q as tar: %s", contextDir, err)
			return
		}

		if err := w.Close(); err != nil {
			errCh <- fmt.Errorf("unable to close buffered piped writer for context dir %q: %s", contextDir, err)
			return
		}
	})

	go func() {
		imageID, err := b.BuildFromDockerfile(ctx, dockerfileData, buildah.BuildFromDockerfileOpts{
			ContextTar: contextTar,
			CommonOpts: buildah.CommonOpts{
				LogWriter: os.Stdout,
			},
		})
		if err != nil {
			errCh <- fmt.Errorf("BuildFromDockerfile failed: %s", err)
			return
		}

		buildDoneCh <- imageID
		close(buildDoneCh)
	}()

	var imageID string
	select {
	case err := <-errCh:
		close(errCh)
		return err

	case imageID = <-buildDoneCh:
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
