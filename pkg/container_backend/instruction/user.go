package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type User struct {
	User string
}

func NewUser(user string) *User {
	return &User{User: user}
}

func (i *User) UsesBuildContext() bool {
	return false
}

func (i *User) Name() string {
	return "USER"
}

func (i *User) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, User: i.User}); err != nil {
		return fmt.Errorf("error setting user %s for container %s: %w", i.User, containerName, err)
	}
	return nil
}
