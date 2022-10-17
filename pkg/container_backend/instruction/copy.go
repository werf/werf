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
	return true
}

func (i *Copy) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	buildContextTmpDir, err := buildContextArchive.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return fmt.Errorf("unable to extract build context: %w", err)
	}

	if err := drv.Copy(ctx, containerName, buildContextTmpDir, i.Src, i.Dst, buildah.CopyOpts{CommonOpts: drvOpts, From: i.From}); err != nil {
		return fmt.Errorf("error copying %v to %s for container %s: %w", i.Src, i.Dst, containerName, err)
	}
	return nil
}
