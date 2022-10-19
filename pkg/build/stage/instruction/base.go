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

	instruction  *dockerfile.DockerfileStageInstruction[T]
	dependencies []*config.Dependency
	hasPrevStage bool
}

func NewBase[T dockerfile.InstructionDataInterface](name stage.StageName, instruction *dockerfile.DockerfileStageInstruction[T], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Base[T] {
	return &Base[T]{
		BaseStage:    stage.NewBaseStage(name, opts),
		instruction:  instruction,
		dependencies: dependencies,
		hasPrevStage: hasPrevStage,
	}
}

func (stg *Base[T]) HasPrevStage() bool {
	return stg.hasPrevStage
}

func (stg *Base[T]) IsStapelStage() bool {
	return false
}

func (stg *Base[T]) prepareInstruction(ctx context.Context, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver, backendInstruction container_backend.InstructionInterface) error {
	stageImage.Builder.DockerfileStageBuilder().SetBuildContextArchive(buildContextArchive) // FIXME(staged-dockerfile): set context at build-phase level
	stageImage.Builder.DockerfileStageBuilder().AppendInstruction(backendInstruction)
	return nil
}
