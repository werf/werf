package instruction

import (
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
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

func (stage *Base[T]) HasPrevStage() bool {
	return stage.hasPrevStage
}

func (s *Base[T]) IsStapelStage() bool {
	return false
}
