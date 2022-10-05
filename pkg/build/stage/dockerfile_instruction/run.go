package dockerfile_instruction

import (
	"context"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/util"
)

type Run struct {
	*Base
	instruction *dockerfile.InstructionRun
}

func NewRun(instruction *dockerfile.InstructionRun, dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Run {
	return &Run{
		Base:        NewBase(stage.StageName(instruction.Name()), dependencies, hasPrevStage, opts),
		instruction: instruction,
	}
}

func (stage *Run) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage) (string, error) {
	return util.Sha256Hash(stage.instruction.Command...), nil
}

func (stage *Run) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage) error {
	stageImage.Builder.DockerfileStageBuilder().AppendMainCommands(stage.instruction)
	return nil
}
