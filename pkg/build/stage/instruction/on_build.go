package instruction

import (
	"context"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	backend_instruction "github.com/werf/werf/v2/pkg/container_backend/instruction"
	"github.com/werf/werf/v2/pkg/dockerfile"
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
	args = stg.addImageCacheVersionToDependencies(args)

	return util.Sha256Hash(args...), nil
}
