package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Copy struct {
	dockerfile_instruction.Copy
}

func NewCopy(i dockerfile_instruction.Copy) *Copy {
	return &Copy{Copy: i}
}

func (i *Copy) UsesBuildContext() bool {
	return i.From == ""
}

func (i *Copy) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	var contextDir string
	if i.UsesBuildContext() {
		var err error
		contextDir, err = buildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return fmt.Errorf("unable to extract build context: %w", err)
		}
	} else {
		container, err := drv.FromCommand(ctx, "", i.From, buildah.FromCommandOpts{})
		if err != nil {
			return fmt.Errorf("unable to create container from image %q: %w", i.From, err)
		}

		contextDir, err = drv.Mount(ctx, container, buildah.MountOpts{})
		if err != nil {
			return fmt.Errorf("unable to mount container %q: %w", container, err)
		}
	}

	if err := drv.Copy(ctx, containerName, contextDir, i.Src, i.Dst, buildah.CopyOpts{
		CommonOpts: drvOpts,
		Chown:      i.Chown,
		Chmod:      i.Chmod,
	}); err != nil {
		return fmt.Errorf("error copying %v to %s for container %s: %w", i.Src, i.Dst, containerName, err)
	}

	return nil
}
