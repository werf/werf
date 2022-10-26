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
	*Base[*dockerfile_instruction.Healthcheck, *backend_instruction.Healthcheck]
}

func NewHealthcheck(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Healthcheck], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Healthcheck {
	return &Healthcheck{Base: NewBase(name, i, backend_instruction.NewHealthcheck(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Healthcheck) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, "Type", string(stg.instruction.Data.Type))
	args = append(args, append([]string{"Test"}, stg.instruction.Data.Config.Test...)...)
	args = append(args, "Interval", stg.instruction.Data.Config.Interval.String())
	args = append(args, "Timeout", stg.instruction.Data.Config.Timeout.String())
	args = append(args, "StartPeriod", stg.instruction.Data.Config.StartPeriod.String())
	args = append(args, "Retries", fmt.Sprintf("%d", stg.instruction.Data.Config.Retries))
	return util.Sha256Hash(args...), nil
}
