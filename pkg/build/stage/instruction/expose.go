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

type Expose struct {
	*Base[*dockerfile_instruction.Expose, *backend_instruction.Expose]
}

func NewExpose(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Expose], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Expose {
	return &Expose{Base: NewBase(name, i, backend_instruction.NewExpose(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Expose) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, append([]string{"Ports"}, stg.instruction.Data.Ports...)...)
	return util.Sha256Hash(args...), nil
}
