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

type Env struct {
	*Base[*dockerfile_instruction.Env, *backend_instruction.Env]
}

func NewEnv(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Env], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Env {
	return &Env{Base: NewBase(name, i, backend_instruction.NewEnv(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Env) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	// FIXME(staged-dockerfile): sort envs
	if len(stg.instruction.Data.Envs) > 0 {
		args = append(args, "Envs")
		for k, v := range stg.instruction.Data.Envs {
			args = append(args, k, v)
		}
	}
	return util.Sha256Hash(args...), nil
}
