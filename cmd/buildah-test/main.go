package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/werf/werf/pkg/docker"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/werf"
)

var errUsage = errors.New("./buildah-test {auto|native-rootless|docker-with-fuse} DOCKERFILE_PATH [CONTEXT_PATH]")

func do(ctx context.Context) error {
	var mode buildah.Mode
	if v := os.Getenv("BUILDAH_TEST_MODE"); v != "" {
		mode = buildah.Mode(v)
	} else {
		if len(os.Args) < 2 {
			return errUsage
		}
		mode = buildah.ResolveMode(buildah.Mode(os.Args[1]))

		os.Setenv("BUILDAH_TEST_MODE", string(mode))
	}

	shouldTerminate, err := buildah.ProcessStartupHook(mode)
	if err != nil {
		return fmt.Errorf("buildah process startup hook failed: %s", err)
	}
	if shouldTerminate {
		return nil
	}

	if err := werf.Init("", ""); err != nil {
		return fmt.Errorf("unable to init werf subsystem: %s", err)
	}

	mode = buildah.ResolveMode(mode)

	fmt.Printf("Using buildah mode: %s\n", mode)

	if mode == buildah.ModeDockerWithFuse {
		if err := docker.Init(ctx, "", false, false, ""); err != nil {
			return err
		}
	}

	if len(os.Args) < 3 {
		return errUsage
	}

	dockerfilePath := os.Args[2]

	var contextDir string
	if len(os.Args) > 3 {
		contextDir = os.Args[3]
	}

	b, err := buildah.NewBuildah(mode, buildah.BuildahOpts{})
	if err != nil {
		return fmt.Errorf("unable to create buildah client: %s", err)
	}

	dockerfileData, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("error reading %q: %s", dockerfilePath, err)
	}

	errCh := make(chan error, 0)
	buildDoneCh := make(chan string, 0)

	var contextTar io.Reader
	if contextDir != "" {
		contextTar = util.BufferedPipedWriterProcess(func(w io.WriteCloser) {
			if err := util.WriteDirAsTar((contextDir), w); err != nil {
				errCh <- fmt.Errorf("unable to write dir %q as tar: %s", contextDir, err)
				return
			}

			if err := w.Close(); err != nil {
				errCh <- fmt.Errorf("unable to close buffered piped writer for context dir %q: %s", contextDir, err)
				return
			}
		})
	}

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

	select {
	case err := <-errCh:
		close(errCh)
		return err

	case imageID := <-buildDoneCh:
		fmt.Fprintf(os.Stdout, "INFO: built imageId is %s\n", imageID)
	}

	return nil
}

func main() {
	if err := do(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
