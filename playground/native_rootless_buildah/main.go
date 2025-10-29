package main

import (
	"context"
	"fmt"
	"github.com/containers/storage/pkg/reexec"
	"log"
	"os"
	"path/filepath"

	"github.com/containers/storage/pkg/unshare"
	"github.com/sirupsen/logrus"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/werf"
)

const newImage = "ilyalesikov/test:test"

func init() {
	logrus.SetLevel(logrus.TraceLevel)

	if reexec.Init() {
		return
	}

	result, err := unshare.HasCapSysAdmin()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("has_cap_sys_admin=%t", result)

	unshare.MaybeReexecUsingUserNamespace(false)
}

func main() {
	ctx := context.TODO()

	if err := werf.Init("", ""); err != nil {
		log.Fatal(err)
	}

	b, err := buildah.NewBuildah(buildah.ModeNative, buildah.BuildahOpts{})
	if err != nil {
		log.Fatal(err)
	}

	if err = runCommand(ctx, b); err != nil {
		log.Fatal(err)
	}

	imageId, err := buildFromDockerfile(ctx, b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "INFO: imageId is %s\n", imageId)

	if err := b.Pull(ctx, "ubuntu:20.04", buildah.PullOpts{LogWriter: os.Stdout}); err != nil {
		log.Fatal(err)
	}

	if err := b.Tag(ctx, "ubuntu:20.04", newImage, buildah.TagOpts{LogWriter: os.Stdout}); err != nil {
		log.Fatal(err)
	}

	builderInfo, err := b.Inspect(ctx, "ubuntu:20.04")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(builderInfo)

	builderInfo, err = b.Inspect(ctx, newImage)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(builderInfo)

	if err := b.Push(ctx, newImage, buildah.PushOpts{LogWriter: os.Stdout}); err != nil {
		log.Fatal(err)
	}

	if err := b.Rmi(ctx, "ubuntu:20.04", buildah.RmiOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: os.Stdout,
		},
	}); err != nil {
		log.Fatal(err)
	}
}

func runCommand(ctx context.Context, b buildah.Buildah) error {
	return b.RunCommand(ctx, "build-container", []string{"ls"}, buildah.RunCommandOpts{})
}

func buildFromDockerfile(ctx context.Context, b buildah.Buildah) (string, error) {
	dockerfileContent := `
FROM ubuntu:20.04
RUN apt update
COPY . /app
`

	imageId, err := b.BuildFromDockerfile(ctx, dockerfileContent, buildah.BuildFromDockerfileOpts{
		ContextDir: filepath.Join(os.Getenv("PWD"), "playground/native_rootless_buildah"),
		CommonOpts: buildah.CommonOpts{
			LogWriter: os.Stdout,
		},
	})
	if err != nil {
		return "", fmt.Errorf("BuildFromDockerfile failed: %w", err)
	}

	return imageId, nil
}
