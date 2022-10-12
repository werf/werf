package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Run struct {
	dockerfile_instruction.Run
}

func NewRun(i dockerfile_instruction.Run) *Run {
	return &Run{Run: i}
}

func (i *Run) UsesBuildContext() bool {
	return false
}

func (i *Run) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.RunCommand(ctx, containerName, i.Command, buildah.RunCommandOpts{
		// FIXME(ilya-lesikov): should we suppress or not?
		CommonOpts: drvOpts,
	}); err != nil {
		return fmt.Errorf("error running command %v for container %s: %w", i.Command, containerName, err)
	}
	return nil
}
