package container_runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/werf/werf/pkg/docker"
)

type DockerfileImageBuilder struct {
	temporalId string
	isBuilt    bool
	BuildArgs  []string
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
	b.BuildArgs = append(b.BuildArgs, buildArgs...)
}

func (b *DockerfileImageBuilder) Build(ctx context.Context) error {
	buildArgs := append(b.BuildArgs, fmt.Sprintf("--tag=%s", b.temporalId))

	if err := docker.CliBuild_LiveOutput(ctx, buildArgs...); err != nil {
		return err
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
