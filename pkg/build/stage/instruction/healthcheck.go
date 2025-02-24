package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	backend_instruction "github.com/werf/werf/v2/pkg/container_backend/instruction"
	"github.com/werf/werf/v2/pkg/dockerfile"
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
