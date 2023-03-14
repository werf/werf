package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Add struct {
	instructions.AddCommand
}

func NewAdd(i instructions.AddCommand) *Add {
	return &Add{AddCommand: i}
}

func (i *Add) UsesBuildContext() bool {
	return true
}

func (i *Add) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	var contextDir string
	if i.UsesBuildContext() {
		var err error
		contextDir, err = buildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return fmt.Errorf("unable to extract build context: %w", err)
		}
	}

	if err := drv.Add(ctx, containerName, i.SourcePaths, i.DestPath, buildah.AddOpts{
		CommonOpts: drvOpts,
		ContextDir: contextDir,
		Chown:      i.Chown,
		Chmod:      i.Chmod,
	}); err != nil {
		return fmt.Errorf("error adding %v to %s for container %s: %w", i.SourcePaths, i.DestPath, containerName, err)
	}

	return nil
}
