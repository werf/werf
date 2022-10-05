package dockerfile_instruction

import (
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
)

type Base struct {
	*stage.BaseStage

	dependencies []*config.Dependency
	hasPrevStage bool
}

func NewBase(name stage.StageName, dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Base {
	return &Base{
		BaseStage:    stage.NewBaseStage(name, opts),
		dependencies: dependencies,
		hasPrevStage: hasPrevStage,
	}
}

func (stage *Base) HasPrevStage() bool {
	return stage.hasPrevStage
}

func (s *Base) IsStapelStage() bool {
	return false
}
