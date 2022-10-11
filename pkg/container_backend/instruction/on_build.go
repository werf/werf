package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type OnBuild struct {
	dockerfile_instruction.OnBuild
}

func NewOnBuild(i dockerfile_instruction.OnBuild) *OnBuild {
	return &OnBuild{OnBuild: i}
}

func (i *OnBuild) UsesBuildContext() bool {
	return false
}

func (i *OnBuild) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, OnBuild: i.Instruction}); err != nil {
		return fmt.Errorf("error setting onbuild %v for container %s: %w", i.Instruction, containerName, err)
	}
	return nil
}
