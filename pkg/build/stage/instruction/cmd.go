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

type Cmd struct {
	*Base[*instructions.CmdCommand, *backend_instruction.Cmd]
}

func NewCmd(i *dockerfile.DockerfileStageInstruction[*instructions.CmdCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Cmd {
	return &Cmd{Base: NewBase(i, backend_instruction.NewCmd(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Cmd) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Cmd"}, stg.instruction.Data.CmdLine...)...)
	args = append(args, "PrependShell", fmt.Sprintf("%v", stg.instruction.Data.PrependShell))

	return util.Sha256Hash(args...), nil
}
