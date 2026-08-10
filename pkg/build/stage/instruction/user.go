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

type User struct {
	*Base[*instructions.UserCommand, *backend_instruction.User]
}

func NewUser(i *dockerfile.DockerfileStageInstruction[*instructions.UserCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *User {
	return &User{Base: NewBase(i, backend_instruction.NewUser(*i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *User) ExpandDependencies(ctx context.Context, c stage.Conveyor, baseEnv map[string]string) error {
	return stg.doExpandDependencies(ctx, c, baseEnv, stg)
}

func (stg *User) ExpandInstruction(c stage.Conveyor, env map[string]string) error {
	if err := stg.Base.ExpandInstruction(c, env); err != nil {
		return err
	}

	stg.backendInstruction.UserCommand = *stg.instruction.Data

	return nil
}

func (stg *User) GetContentDependencies(ctx context.Context, c stage.Conveyor, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	return stg.GetDependencies(ctx, c, nil, nil, nil, buildContextArchive)
}

func (stg *User) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "User", stg.instruction.Data.User)

	return util.Sha256Hash(args...), nil
}
