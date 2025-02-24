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
