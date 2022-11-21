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

type OnBuild struct {
	*Base[*instructions.OnbuildCommand, *backend_instruction.OnBuild]
}

func NewOnBuild(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*instructions.OnbuildCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *OnBuild {
	return &OnBuild{Base: NewBase(name, i, backend_instruction.NewOnBuild(i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *OnBuild) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, "Expression", stg.instruction.Data.Expression)
	return util.Sha256Hash(args...), nil
}
