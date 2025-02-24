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

type Env struct {
	*Base[*instructions.EnvCommand, *backend_instruction.Env]
}

func NewEnv(i *dockerfile.DockerfileStageInstruction[*instructions.EnvCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Env {
	return &Env{Base: NewBase(i, backend_instruction.NewEnv(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Env) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	if len(stg.instruction.Data.Env) > 0 {
		args = append(args, "Env")
		for _, item := range stg.instruction.Data.Env {
			args = append(args, item.Key, item.Value)
		}
	}

	return util.Sha256Hash(args...), nil
}
