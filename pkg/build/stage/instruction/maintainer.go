package instruction

import (
	"context"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

type Maintainer struct {
	*Base[*dockerfile_instruction.Maintainer]
}

func NewMaintainer(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Maintainer], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Maintainer {
	// FIXME(staged-dockerfile): no Maintainer instruction
	return &Maintainer{Base: NewBase(name, i, nil, dependencies, hasPrevStage, opts)}
}

func (stage *Maintainer) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.Maintainer)
	return util.Sha256Hash(args...), nil
}
