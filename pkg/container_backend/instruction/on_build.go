package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type OnBuild struct {
	Instruction string
}

func NewOnBuild(instruction string) *OnBuild {
	return &OnBuild{Instruction: instruction}
}

func (i *OnBuild) UsesBuildContext() bool {
	return false
}

func (i *OnBuild) Name() string {
	return "ONBUILD"
}

func (i *OnBuild) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, OnBuild: i.Instruction}); err != nil {
		return fmt.Errorf("error setting onbuild %v for container %s: %w", i.Instruction, containerName, err)
	}
	return nil
}
