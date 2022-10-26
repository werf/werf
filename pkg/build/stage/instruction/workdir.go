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
	*Base[*dockerfile_instruction.Workdir, *backend_instruction.Workdir]
}

func NewWorkdir(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Workdir], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Workdir {
	return &Workdir{Base: NewBase(name, i, backend_instruction.NewWorkdir(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Workdir) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, "Workdir", stg.instruction.Data.Workdir)
	return util.Sha256Hash(args...), nil
}
