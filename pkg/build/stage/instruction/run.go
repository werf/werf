package instruction

import (
	"context"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

type Run struct {
	*Base[*dockerfile_instruction.Run]
}

func NewRun(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Run], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Run {
	return &Run{Base: NewBase(name, i, dependencies, hasPrevStage, opts)}
}

func (stage *Run) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	return util.Sha256Hash(append([]string{stage.instruction.Data.Name()}, stage.instruction.Data.Command...)...), nil
}

func (stage *Run) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	return stage.prepareInstruction(ctx, stageImage, buildContextArchive, backend_instruction.NewRun(*stage.instruction.Data))
}
