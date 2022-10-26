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

type Cmd struct {
	*Base[*dockerfile_instruction.Cmd, *backend_instruction.Cmd]
}

func NewCmd(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Cmd], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Cmd {
	return &Cmd{Base: NewBase(name, i, backend_instruction.NewCmd(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Cmd) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, append([]string{"Cmd"}, stg.instruction.Data.Cmd...)...)
	args = append(args, "PrependShell", fmt.Sprintf("%v", stg.instruction.Data.PrependShell))

	return util.Sha256Hash(args...), nil
}
