package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/containers/storage/pkg/reexec"
	"github.com/containers/storage/pkg/unshare"
	"github.com/sirupsen/logrus"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/werf"
)

const newImage = "ilyalesikov/test:test"

func init() {
	logrus.SetLevel(logrus.TraceLevel)

	unshare.MaybeReexecUsingUserNamespace(false)
}

func main() {
	if reexec.Init() {
		return
	}

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
	contextFile := filepath.Join(os.Getenv("HOME"), "go", "src", "github.com", "werf", "werf", "playground", "native_buildah", "context.tar")
	tarFileReader, err := os.OpenFile(contextFile, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer tarFileReader.Close()

	imageId, err := b.BuildFromDockerfile(context.Background(),
		[]byte(`FROM ubuntu:20.04
RUN apt update
COPY . /app
`), buildah.BuildFromDockerfileOpts{
			ContextTar: tarFileReader,
			CommonOpts: buildah.CommonOpts{
				LogWriter: os.Stdout,
			},
		})
	if err != nil {
		return "", fmt.Errorf("BuildFromDockerfile failed: %w", err)
	}

	return imageId, nil
}
