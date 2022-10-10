package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Add struct {
	Src []string
	Dst string
}

func NewAdd(src []string, dst string) *Add {
	return &Add{Src: src, Dst: dst}
}

func (i *Add) UsesBuildContext() bool {
	return true
}

func (i *Add) Name() string {
	return "ADD"
}

func (i *Add) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	contextDir, err := buildContext.GetContextDir(ctx)
	if err != nil {
		return fmt.Errorf("unable to get build context dir: %w", err)
	}

	if err := drv.Add(ctx, containerName, i.Src, i.Dst, buildah.AddOpts{CommonOpts: drvOpts, ContextDir: contextDir}); err != nil {
		return fmt.Errorf("error adding %v to %s for container %s: %w", i.Src, i.Dst, containerName, err)
	}
	return nil
}
