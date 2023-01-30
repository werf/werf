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

type Label struct {
	*Base[*instructions.LabelCommand, *backend_instruction.Label]
}

func NewLabel(i *dockerfile.DockerfileStageInstruction[*instructions.LabelCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Label {
	return &Label{Base: NewBase(i, backend_instruction.NewLabel(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Label) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	if len(stg.instruction.Data.Labels) > 0 {
		args = append(args, "Labels")
		for _, item := range stg.instruction.Data.Labels {
			args = append(args, item.Key, item.Value)
		}
	}

	return util.Sha256Hash(args...), nil
}
