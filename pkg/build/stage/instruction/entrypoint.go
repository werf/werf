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

type Entrypoint struct {
	*Base[*dockerfile_instruction.Entrypoint]
}

func NewEntrypoint(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Entrypoint], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Entrypoint {
	return &Entrypoint{Base: NewBase(name, i, backend_instruction.NewEntrypoint(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stage *Entrypoint) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	args = append(args, stage.instruction.Data.Entrypoint...)
	args = append(args, fmt.Sprintf("%v", stage.instruction.Data.PrependShell))
	return util.Sha256Hash(args...), nil
}
