package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type User struct {
	instructions.UserCommand
}

func NewUser(i instructions.UserCommand) *User {
	return &User{UserCommand: i}
}

func (i *User) UsesBuildContext() bool {
	return false
}

func (i *User) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, User: i.User}); err != nil {
		return fmt.Errorf("error setting user %s for container %s: %w", i.User, containerName, err)
	}
	return nil
}
