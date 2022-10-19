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

type Volume struct {
	*Base[*dockerfile_instruction.Volume]
}

func NewVolume(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Volume], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Volume {
	return &Volume{Base: NewBase(name, i, backend_instruction.NewVolume(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stage *Volume) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.Volumes...)
	return util.Sha256Hash(args...), nil
}
