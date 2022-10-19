package instruction

import (
	"context"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/dockerfile"
)

type Base[T dockerfile.InstructionDataInterface] struct {
	*stage.BaseStage

	instruction        *dockerfile.DockerfileStageInstruction[T]
	backendInstruction container_backend.InstructionInterface
	dependencies       []*config.Dependency
	hasPrevStage       bool
}

func NewBase[T dockerfile.InstructionDataInterface](name stage.StageName, instruction *dockerfile.DockerfileStageInstruction[T], backendInstruction container_backend.InstructionInterface, dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Base[T] {
	return &Base[T]{
		BaseStage:          stage.NewBaseStage(name, opts),
		instruction:        instruction,
		backendInstruction: backendInstruction,
		dependencies:       dependencies,
		hasPrevStage:       hasPrevStage,
	}
}

func (stg *Base[T]) HasPrevStage() bool {
	return stg.hasPrevStage
}

func (stg *Base[T]) IsStapelStage() bool {
	return false
}

func (stg *Base[T]) UsesBuildContext() bool {
	return stg.backendInstruction.UsesBuildContext()
}

func (stg *Base[T]) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	stageImage.Builder.DockerfileStageBuilder().AppendInstruction(stg.backendInstruction)
	return nil
}
