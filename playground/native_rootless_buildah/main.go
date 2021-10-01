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

	b, err := buildah.NewBuildah(buildah.ModeNativeRootless, buildah.BuildahOpts{})
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

	if err := b.Pull(context.Background(), "alpine:3.14.2", buildah.PullOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: os.Stdout,
		},
	}); err != nil {
		log.Fatal(err)
	}

	if err := b.Tag(context.Background(), "alpine:3.14.2", "ilyalesikov/test:test", buildah.TagOpts{
		CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout},
	}); err != nil {
		log.Fatal(err)
	}

	builderInfo, err := b.Inspect(context.Background(), "ilyalesikov/test:test")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(builderInfo)

	if err := b.Push(context.Background(), "ilyalesikov/test:test", buildah.PushOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: os.Stdout,
		},
	}); err != nil {
		log.Fatal(err)
	}

	if err := b.Rmi(context.Background(), "alpine:3.14.2", buildah.RmiOpts{
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
	contextFile := filepath.Join(os.Getenv("HOME"), ".go", "src", "github.com", "werf", "werf", "playground", "buildah", "context.tar")
	tarFileReader, err := os.OpenFile(contextFile, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer tarFileReader.Close()

	imageId, err := b.BuildFromDockerfile(context.Background(),
		[]byte(`FROM alpine
RUN wget ya.ru
COPY . /app
`), buildah.BuildFromDockerfileOpts{
			ContextTar: tarFileReader,
			CommonOpts: buildah.CommonOpts{
				LogWriter: os.Stdout,
			},
		})
	if err != nil {
		return "", fmt.Errorf("BuildFromDockerfile failed: %s", err)
	}

	return imageId, nil
}
