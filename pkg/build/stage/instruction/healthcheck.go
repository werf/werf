package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

type Healthcheck struct {
	*Base[*dockerfile_instruction.Healthcheck]
}

func NewHealthcheck(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Healthcheck], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Healthcheck {
	return &Healthcheck{Base: NewBase(name, i, backend_instruction.NewHealthcheck(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stage *Healthcheck) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, string(stage.instruction.Data.Type))
	args = append(args, stage.instruction.Data.Config.Test...)
	args = append(args, stage.instruction.Data.Config.Interval.String())
	args = append(args, stage.instruction.Data.Config.Timeout.String())
	args = append(args, stage.instruction.Data.Config.StartPeriod.String())
	args = append(args, fmt.Sprintf("%d", stage.instruction.Data.Config.Retries))
	return util.Sha256Hash(args...), nil
}
