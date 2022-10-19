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

type OnBuild struct {
	*Base[*dockerfile_instruction.OnBuild]
}

func NewOnBuild(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.OnBuild], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *OnBuild {
	return &OnBuild{Base: NewBase(name, i, dependencies, hasPrevStage, opts)}
}

func (stage *OnBuild) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.Instruction)
	return util.Sha256Hash(args...), nil
}

func (stage *OnBuild) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	return stage.Base.prepareInstruction(ctx, stageImage, buildContextArchive, backend_instruction.NewOnBuild(*stage.instruction.Data))
}
