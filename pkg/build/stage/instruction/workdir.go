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

type Workdir struct {
	*Base[*instructions.WorkdirCommand, *backend_instruction.Workdir]
}

func NewWorkdir(i *dockerfile.DockerfileStageInstruction[*instructions.WorkdirCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Workdir {
	return &Workdir{Base: NewBase(i, backend_instruction.NewWorkdir(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Workdir) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "Path", stg.instruction.Data.Path)

	return util.Sha256Hash(args...), nil
}
