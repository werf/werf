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

type Volume struct {
	*Base[*instructions.VolumeCommand, *backend_instruction.Volume]
}

func NewVolume(i *dockerfile.DockerfileStageInstruction[*instructions.VolumeCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Volume {
	return &Volume{Base: NewBase(i, backend_instruction.NewVolume(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Volume) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Volumes"}, stg.instruction.Data.Volumes...)...)

	return util.Sha256Hash(args...), nil
}
