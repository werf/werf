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

type Shell struct {
	*Base[*dockerfile_instruction.Shell, *backend_instruction.Shell]
}

func NewShell(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Shell], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Shell {
	return &Shell{Base: NewBase(name, i, backend_instruction.NewShell(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Shell) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, append([]string{"Shell"}, stg.instruction.Data.Shell...)...)
	return util.Sha256Hash(args...), nil
}
