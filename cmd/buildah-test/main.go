package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/werf"
)

var errUsage = errors.New(`
    ./buildah-test {auto|native} { dockerfile DOCKERFILE_PATH [CONTEXT_PATH] |
                                                    stapel }
`)

func runStapel(ctx context.Context, mode buildah.Mode) error {
	b, err := buildah.NewBuildah(mode, buildah.BuildahOpts{})
	if err != nil {
		return fmt.Errorf("unable to create buildah client: %w", err)
	}

	// TODO: b.Rm(ctx, "mycontainer", buildah.RmOpts{})

	if _, err := b.FromCommand(ctx, "mycontainer", "ubuntu:20.04", buildah.FromCommandOpts{}); err != nil {
		return fmt.Errorf("unable to create mycontainer from ubuntu:20.04: %w", err)
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
		GlobalMounts: []*specs.Mount{
			{
				Type:        "bind",
				Source:      "/tmp/build_stage.sh",
				Destination: "/.werf/build_stage.sh",
			},
		},
	}); err != nil {
		return fmt.Errorf("unable to run build_stage.sh: %w", err)
	}

	containerRootDir, err := b.Mount(ctx, "mycontainer", buildah.MountOpts{})
	if err != nil {
		return fmt.Errorf("unable to mount mycontainer root dir: %w", err)
	}
	defer b.Umount(ctx, containerRootDir, buildah.UmountOpts{})

	if err := os.WriteFile(filepath.Join(containerRootDir, "/FILE_FROM_GOLANG"), []byte("HELLO WORLD\n"), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write /FILE_FROM_GOLANG into %q: %w", containerRootDir, err)
	}

	// TODO: b.Commit(ctx, "mycontainer, "docker://ghcr.io/GROUP/NAME:TAG", buildah.CommitOpts{})

	return nil
}

func runDockerfile(ctx context.Context, mode buildah.Mode, dockerfilePath, contextDir string) error {
	b, err := buildah.NewBuildah(mode, buildah.BuildahOpts{})
	if err != nil {
		return fmt.Errorf("unable to create buildah client: %w", err)
	}

	errCh := make(chan error, 0)
	buildDoneCh := make(chan string, 0)

	go func() {
		imageID, err := b.BuildFromDockerfile(ctx, dockerfilePath, buildah.BuildFromDockerfileOpts{
			ContextDir: contextDir,
			CommonOpts: buildah.CommonOpts{
				LogWriter: os.Stdout,
			},
		})
		if err != nil {
			errCh <- fmt.Errorf("BuildFromDockerfile failed: %w", err)
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
		mode = buildah.Mode(os.Args[1])

		os.Setenv("BUILDAH_TEST_MODE", string(mode))
	}

	shouldTerminate, err := buildah.ProcessStartupHook(mode)
	if err != nil {
		return fmt.Errorf("buildah process startup hook failed: %w", err)
	}
	if shouldTerminate {
		return nil
	}

	if err := werf.Init("", ""); err != nil {
		return fmt.Errorf("unable to init werf subsystem: %w", err)
	}

	fmt.Printf("Using buildah mode: %s\n", mode)

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
		return fmt.Errorf("bad argument given %q: expected dockerfile or stapel\n\n%w\n", os.Args[2], errUsage)
	}
}

func main() {
	if err := do(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
