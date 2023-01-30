package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/util"
)

type Healthcheck struct {
	*Base[*instructions.HealthCheckCommand, *backend_instruction.Healthcheck]
}

func NewHealthcheck(i *dockerfile.DockerfileStageInstruction[*instructions.HealthCheckCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Healthcheck {
	return &Healthcheck{Base: NewBase(i, backend_instruction.NewHealthcheck(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Healthcheck) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Test"}, stg.instruction.Data.Health.Test...)...)
	args = append(args, "Interval", stg.instruction.Data.Health.Interval.String())
	args = append(args, "Timeout", stg.instruction.Data.Health.Timeout.String())
	args = append(args, "StartPeriod", stg.instruction.Data.Health.StartPeriod.String())
	args = append(args, "Retries", fmt.Sprintf("%d", stg.instruction.Data.Health.Retries))

	return util.Sha256Hash(args...), nil
}
