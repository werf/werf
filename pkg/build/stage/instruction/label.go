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

type Label struct {
	*Base[*dockerfile_instruction.Label]
}

func NewLabel(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Label], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Label {
	return &Label{Base: NewBase(name, i, backend_instruction.NewLabel(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stage *Label) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, stage.instruction.Data.Name())
	// FIXME(staged-dockerfile): sort labels map
	for k, v := range stage.instruction.Data.Labels {
		args = append(args, k, v)
	}
	return util.Sha256Hash(args...), nil
}
