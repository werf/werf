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

type Shell struct {
	*Base[*instructions.ShellCommand, *backend_instruction.Shell]
}

func NewShell(i *dockerfile.DockerfileStageInstruction[*instructions.ShellCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Shell {
	return &Shell{Base: NewBase(i, backend_instruction.NewShell(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Shell) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Shell"}, stg.instruction.Data.Shell...)...)
	args = stg.addImageCacheVersionToDependencies(args)

	return util.Sha256Hash(args...), nil
}
