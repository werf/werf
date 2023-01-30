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

func NewOnBuild(i *dockerfile.DockerfileStageInstruction[*instructions.OnbuildCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *OnBuild {
	return &OnBuild{Base: NewBase(i, backend_instruction.NewOnBuild(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *OnBuild) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "Expression", stg.instruction.Data.Expression)

	return util.Sha256Hash(args...), nil
}
