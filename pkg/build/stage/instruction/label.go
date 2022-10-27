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
	*Base[*dockerfile_instruction.Label, *backend_instruction.Label]
}

func NewLabel(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Label], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Label {
	return &Label{Base: NewBase(name, i, backend_instruction.NewLabel(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Label) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	if len(stg.instruction.Data.Labels) > 0 {
		args = append(args, "Labels")
		for _, k := range util.SortedStringKeys(stg.instruction.Data.Labels) {
			args = append(args, k, stg.instruction.Data.Labels[k])
		}
	}
	return util.Sha256Hash(args...), nil
}
