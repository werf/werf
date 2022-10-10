package instruction

import (
	"context"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/util"
)

type Run struct {
	*Base
	instruction *backend_instruction.Run
}

func NewRun(i *backend_instruction.Run, dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Run {
	return &Run{
		Base:        NewBase(InstructionRun, dependencies, hasPrevStage, opts),
		instruction: i,
	}
}

func (stage *Run) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage) (string, error) {
	return util.Sha256Hash(append([]string{string(InstructionRun)}, stage.instruction.Command...)...), nil
}

func (stage *Run) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage) error {
	stageImage.Builder.DockerfileStageBuilder().AppendInstruction(stage.instruction)
	return nil
}
