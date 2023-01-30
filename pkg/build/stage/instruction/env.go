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
