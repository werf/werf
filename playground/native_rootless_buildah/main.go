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
	if err := werf.Init("", ""); err != nil {
		log.Fatal(err)
	}

	b, err := buildah.NewBuildah(buildah.ModeNative, buildah.BuildahOpts{})
	if err != nil {
		log.Fatal(err)
	}

	if err = runCommand(b); err != nil {
		log.Fatal(err)
	}

	imageId, err := buildFromDockerfile(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "INFO: imageId is %s\n", imageId)

	if err := b.Pull(context.Background(), "ubuntu:20.04", buildah.PullOpts{LogWriter: os.Stdout}); err != nil {
		log.Fatal(err)
	}

	if err := b.Tag(context.Background(), "ubuntu:20.04", newImage, buildah.TagOpts{LogWriter: os.Stdout}); err != nil {
		log.Fatal(err)
	}

	builderInfo, err := b.Inspect(context.Background(), "ubuntu:20.04")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(builderInfo)

	builderInfo, err = b.Inspect(context.Background(), newImage)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(builderInfo)

	if err := b.Push(context.Background(), newImage, buildah.PushOpts{LogWriter: os.Stdout}); err != nil {
		log.Fatal(err)
	}

	if err := b.Rmi(context.Background(), "ubuntu:20.04", buildah.RmiOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: os.Stdout,
		},
	}); err != nil {
		log.Fatal(err)
	}
}

func runCommand(b buildah.Buildah) error {
	return b.RunCommand(context.Background(), "build-container", []string{"ls"}, buildah.RunCommandOpts{})
}

func buildFromDockerfile(b buildah.Buildah) (string, error) {
	dockerfileContent := `
FROM ubuntu:20.04
RUN apt update
COPY . /app
`

	imageId, err := b.BuildFromDockerfile(context.Background(), dockerfileContent, buildah.BuildFromDockerfileOpts{
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
