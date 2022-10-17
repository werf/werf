package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Add struct {
	dockerfile_instruction.Add
}

func NewAdd(i dockerfile_instruction.Add) *Add {
	return &Add{Add: i}
}

func (i *Add) UsesBuildContext() bool {
	return true
}

func (i *Add) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	buildContextTmpDir, err := buildContextArchive.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return fmt.Errorf("unable to extract build context: %w", err)
	}

	if err := drv.Add(ctx, containerName, i.Src, i.Dst, buildah.AddOpts{CommonOpts: drvOpts, ContextDir: buildContextTmpDir}); err != nil {
		return fmt.Errorf("error adding %v to %s for container %s: %w", i.Src, i.Dst, containerName, err)
	}
	return nil
}
