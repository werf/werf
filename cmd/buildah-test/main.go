package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var errUsage = errors.New(`
    ./buildah-test {auto|native|docker-with-fuse} { dockerfile DOCKERFILE_PATH [CONTEXT_PATH] |
                                                    stapel }
`)

func runStapel(ctx context.Context, mode buildah.Mode) error {
	b, err := buildah.NewBuildah(mode, buildah.BuildahOpts{})
	if err != nil {
		return fmt.Errorf("unable to create buildah client: %s", err)
	}

	// TODO: b.Rm(ctx, "mycontainer", buildah.RmOpts{})

	if _, err := b.FromCommand(ctx, "mycontainer", "ubuntu:20.04", buildah.FromCommandOpts{}); err != nil {
		return fmt.Errorf("unable to create mycontainer from ubuntu:20.04: %s", err)
	}

	buildStageSh := `#!/bin/bash

echo START
id
ls -lah /
echo "HELLO FROM BUILD INSTRUCTION" > /root/HELLO
ls -lah /root
echo STOP
`

	if err := os.WriteFile("/tmp/build_stage.sh", []byte(buildStageSh), os.ModePerm); err != nil {
		return err
	}
	defer os.RemoveAll("/tmp/build_stage.sh")

	if err := b.RunCommand(ctx, "mycontainer", []string{"/.werf/build_stage.sh"}, buildah.RunCommandOpts{
		Mounts: []specs.Mount{
			{
				Type:        "bind",
				Source:      "/tmp/build_stage.sh",
				Destination: "/.werf/build_stage.sh",
			},
		},
	}); err != nil {
		return fmt.Errorf("unable to run build_stage.sh: %s", err)
	}

	containerRootDir, err := b.Mount(ctx, "mycontainer", buildah.MountOpts{})
	if err != nil {
		return fmt.Errorf("unable to mount mycontainer root dir: %s", err)
	}
	defer b.Umount(ctx, containerRootDir, buildah.UmountOpts{})

	if err := os.WriteFile(filepath.Join(containerRootDir, "/FILE_FROM_GOLANG"), []byte("HELLO WORLD\n"), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write /FILE_FROM_GOLANG into %q: %s", containerRootDir, err)
	}

	// TODO: b.Commit(ctx, "mycontainer, "docker://ghcr.io/GROUP/NAME:TAG", buildah.CommitOpts{})

	return nil
}

func runDockerfile(ctx context.Context, mode buildah.Mode, dockerfilePath, contextDir string) error {
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

func do(ctx context.Context) error {
	var mode buildah.Mode

	if v := os.Getenv("BUILDAH_TEST_MODE"); v != "" {
		mode = buildah.Mode(v)
	} else if strings.HasPrefix(os.Args[0], "buildah-") || strings.HasPrefix(os.Args[0], "chrootuser-") || strings.HasPrefix(os.Args[0], "storage-") {
		mode = "native"
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

	switch os.Args[2] {
	case "dockerfile":
		if len(os.Args) < 4 {
			return errUsage
		}

		dockerfilePath := os.Args[3]

		var contextDir string
		if len(os.Args) > 4 {
			contextDir = os.Args[4]
		}

		return runDockerfile(ctx, mode, dockerfilePath, contextDir)
	case "stapel":
		return runStapel(ctx, mode)
	default:
		return fmt.Errorf("bad argument given %q: expected dockerfile or stapel\n\n%s\n", os.Args[2], errUsage)
	}
}

func main() {
	if err := do(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
