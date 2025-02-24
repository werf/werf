package instruction

import (
	"context"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	backend_instruction "github.com/werf/werf/v2/pkg/container_backend/instruction"
	"github.com/werf/werf/v2/pkg/dockerfile"
)

type Maintainer struct {
	*Base[*instructions.MaintainerCommand, *backend_instruction.Maintainer]
}

func NewMaintainer(i *dockerfile.DockerfileStageInstruction[*instructions.MaintainerCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Maintainer {
	return &Maintainer{Base: NewBase(i, backend_instruction.NewMaintainer(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Maintainer) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "Maintainer", stg.instruction.Data.Maintainer)

	return util.Sha256Hash(args...), nil
}
