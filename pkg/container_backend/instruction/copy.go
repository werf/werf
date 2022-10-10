package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Copy struct {
	From string
	Src  []string
	Dst  string
}

func NewCopy(from string, src []string, dst string) *Copy {
	return &Copy{From: from, Src: src, Dst: dst}
}

func (i *Copy) UsesBuildContext() bool {
	return true
}

func (i *Copy) Name() string {
	return "COPY"
}

func (i *Copy) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	contextDir, err := buildContext.GetContextDir(ctx)
	if err != nil {
		return fmt.Errorf("unable to get build context dir: %w", err)
	}

	if err := drv.Copy(ctx, containerName, contextDir, i.Src, i.Dst, buildah.CopyOpts{CommonOpts: drvOpts, From: i.From}); err != nil {
		return fmt.Errorf("error copying %v to %s for container %s: %w", i.Src, i.Dst, containerName, err)
	}
	return nil
}
