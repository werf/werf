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

type Copy struct {
	*Base[*dockerfile_instruction.Copy]
}

func NewCopy(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Copy], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Copy {
	return &Copy{Base: NewBase(name, i, backend_instruction.NewCopy(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stage *Copy) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.From)
	args = append(args, stage.instruction.Data.Src...)
	args = append(args, stage.instruction.Data.Dst)
	args = append(args, stage.instruction.Data.Chown)
	args = append(args, stage.instruction.Data.Chmod)
	return util.Sha256Hash(args...), nil
}
