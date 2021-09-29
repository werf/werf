package container_runtime

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/werf/werf/pkg/docker"
)

//type DockerfileImageBuilder struct {
//	ContainerRuntime       ContainerRuntime
//	Dockerfile             []byte
//	BuildDockerfileOptions BuildDockerfileOptions
//
//	builtID string
//}

type DockerfileImageBuilder struct {
	temporalId      string
	isBuilt         bool
	buildArgs       []string
	filePathToStdin string
}

func NewDockerfileImageBuilder() *DockerfileImageBuilder {
	return &DockerfileImageBuilder{temporalId: uuid.New().String()}
}

func (b *DockerfileImageBuilder) GetBuiltId() string {
	if !b.isBuilt {
		return ""
	}
	return b.temporalId
}

func (b *DockerfileImageBuilder) AppendBuildArgs(buildArgs ...string) {
	b.buildArgs = append(b.buildArgs, buildArgs...)
}

func (b *DockerfileImageBuilder) SetFilePathToStdin(path string) {
	b.filePathToStdin = path
}

func (b *DockerfileImageBuilder) Build(ctx context.Context) error {
	buildArgs := append(b.buildArgs, fmt.Sprintf("--tag=%s", b.temporalId))

	if b.filePathToStdin != "" {
		buildArgs = append(buildArgs, "-")

		f, err := os.Open(b.filePathToStdin)
		if err != nil {
			return fmt.Errorf("unable to open file: %s", err)
		}
		defer f.Close()

		if debugDockerRunCommand() {
			fmt.Printf("Docker run command:\ndocker build %s < %s\n", strings.Join(buildArgs, " "), b.filePathToStdin)
		}

		if err := docker.CliBuild_LiveOutputWithCustomIn(ctx, f, buildArgs...); err != nil {
			return err
		}
	} else {
		if debugDockerRunCommand() {
			fmt.Printf("Docker run command:\ndocker build %s\n", strings.Join(buildArgs, " "))
		}

		if err := docker.CliBuild_LiveOutput(ctx, buildArgs...); err != nil {
			return err
		}
	}

	b.isBuilt = true

	return nil
}

func (b *DockerfileImageBuilder) Cleanup(ctx context.Context) error {
	if err := docker.CliRmi(ctx, b.temporalId, "--force"); err != nil {
		return fmt.Errorf("unable to remove temporal dockerfile image %q: %s", b.temporalId, err)
	}
	return nil
}
