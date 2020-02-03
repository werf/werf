package image

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/flant/werf/pkg/docker"
)

type DockerfileImageBuilder struct {
	temporalId string
	isBuilt    bool
	BuildArgs  []string
}

func NewDockerfileImageBuilder() *DockerfileImageBuilder {
	return &DockerfileImageBuilder{temporalId: uuid.New().String()}
}

func (b *DockerfileImageBuilder) GetBuiltId() (string, error) {
	if !b.isBuilt {
		return "", fmt.Errorf("dockerfile image %s not built yet", b.temporalId)
	}
	return b.temporalId, nil
}

func (b *DockerfileImageBuilder) AppendBuildArgs(buildArgs ...string) {
	b.BuildArgs = append(b.BuildArgs, buildArgs...)
}

func (b *DockerfileImageBuilder) Build() error {
	buildArgs := append(b.BuildArgs, fmt.Sprintf("--tag=%s", b.temporalId))

	if err := docker.CliBuild(buildArgs...); err != nil {
		return err
	}

	b.isBuilt = true

	return nil
}

func (b *DockerfileImageBuilder) Cleanup() error {
	// TODO: remove temporalId
	return nil
}
