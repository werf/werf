package instruction

import (
	"context"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/util"
)

type Maintainer struct {
	*Base[*instructions.MaintainerCommand, *backend_instruction.Maintainer]
}

func NewMaintainer(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*instructions.MaintainerCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Maintainer {
	return &Maintainer{Base: NewBase(name, i, backend_instruction.NewMaintainer(i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Maintainer) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, "Maintainer", stg.instruction.Data.Maintainer)
	return util.Sha256Hash(args...), nil
}
