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

type Expose struct {
	*Base[*instructions.ExposeCommand, *backend_instruction.Expose]
}

func NewExpose(i *dockerfile.DockerfileStageInstruction[*instructions.ExposeCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Expose {
	return &Expose{Base: NewBase(i, backend_instruction.NewExpose(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Expose) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Ports"}, stg.instruction.Data.Ports...)...)

	return util.Sha256Hash(args...), nil
}
