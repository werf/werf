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

type Workdir struct {
	*Base[*dockerfile_instruction.Workdir]
}

func NewWorkdir(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Workdir], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Workdir {
	return &Workdir{Base: NewBase(name, i, dependencies, hasPrevStage, opts)}
}

func (stage *Workdir) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.Workdir)
	return util.Sha256Hash(args...), nil
}

func (stage *Workdir) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	return stage.Base.prepareInstruction(ctx, stageImage, buildContextArchive, backend_instruction.NewWorkdir(*stage.instruction.Data))
}
