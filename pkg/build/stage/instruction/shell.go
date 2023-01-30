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

type Shell struct {
	*Base[*instructions.ShellCommand, *backend_instruction.Shell]
}

func NewShell(i *dockerfile.DockerfileStageInstruction[*instructions.ShellCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Shell {
	return &Shell{Base: NewBase(i, backend_instruction.NewShell(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Shell) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Shell"}, stg.instruction.Data.Shell...)...)

	return util.Sha256Hash(args...), nil
}
