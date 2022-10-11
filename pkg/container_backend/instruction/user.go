package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type User struct {
	dockerfile_instruction.User
}

func NewUser(i dockerfile_instruction.User) *User {
	return &User{User: i}
}

func (i *User) UsesBuildContext() bool {
	return false
}

func (i *User) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, User: i.User.User}); err != nil {
		return fmt.Errorf("error setting user %s for container %s: %w", i.User, containerName, err)
	}
	return nil
}
