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

type StopSignal struct {
	*Base[*instructions.StopSignalCommand, *backend_instruction.StopSignal]
}

func NewStopSignal(i *dockerfile.DockerfileStageInstruction[*instructions.StopSignalCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *StopSignal {
	return &StopSignal{Base: NewBase(i, backend_instruction.NewStopSignal(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *StopSignal) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "Signal", stg.instruction.Data.Signal)

	return util.Sha256Hash(args...), nil
}
