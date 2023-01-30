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

type Entrypoint struct {
	*Base[*instructions.EntrypointCommand, *backend_instruction.Entrypoint]
}

func NewEntrypoint(i *dockerfile.DockerfileStageInstruction[*instructions.EntrypointCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Entrypoint {
	return &Entrypoint{Base: NewBase(i, backend_instruction.NewEntrypoint(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Entrypoint) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Entrypoint"}, stg.instruction.Data.CmdLine...)...)
	args = append(args, "PrependShell", fmt.Sprintf("%v", stg.instruction.Data.PrependShell))

	return util.Sha256Hash(args...), nil
}
