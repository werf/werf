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

type Entrypoint struct {
	*Base[*instructions.EntrypointCommand, *backend_instruction.Entrypoint]
}

func NewEntrypoint(i *dockerfile.DockerfileStageInstruction[*instructions.EntrypointCommand], dependencies []*config.Dependency, hasPrevStage, entrypointResetCMD bool, opts *stage.BaseStageOptions) *Entrypoint {
	return &Entrypoint{Base: NewBase(i, backend_instruction.NewEntrypoint(*i.Data, entrypointResetCMD), dependencies, hasPrevStage, opts)}
}

func (stg *Entrypoint) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, append([]string{"Entrypoint"}, stg.instruction.Data.CmdLine...)...)
	args = append(args, "PrependShell", fmt.Sprintf("%v", stg.instruction.Data.PrependShell))
	args = stg.addImageCacheVersionToDependencies(args)

	return util.Sha256Hash(args...), nil
}
