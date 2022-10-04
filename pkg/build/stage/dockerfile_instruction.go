package stage

import (
	"context"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/dockerfile"
)

// TODO(staged-dockerfile): common implementation and possible implementation for each separate dockerfile instruction

type DockerfileInstruction interface {
	Name() string
}

type DockerfileInstructionStage struct {
	*BaseStage

	instruction  DockerfileInstruction
	dependencies []*config.Dependency
}

func NewDockerfileInstructionStage(instruction DockerfileInstruction, dependencies []*config.Dependency, dockerfileStage *dockerfile.DockerfileStage, opts *NewBaseStageOptions) *DockerfileInstructionStage {
	return &DockerfileInstructionStage{
		instruction:  instruction,
		dependencies: dependencies,
		BaseStage:    newBaseStage(StageName(instruction.Name()), opts),
	}
}

func (stage *DockerfileInstructionStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage) (string, error) {
	// TODO: digest
	return "", nil
}

func (stage *DockerfileInstructionStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage) error {
	// TODO: setup builder
	return nil
}
