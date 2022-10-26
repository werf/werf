package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Run struct {
	dockerfile_instruction.Run
}

func NewRun(i dockerfile_instruction.Run) *Run {
	return &Run{Run: i}
}

func (i *Run) UsesBuildContext() bool {
	for _, mount := range i.Mounts {
		if mount.Type == instructions.MountTypeBind && mount.From == "" {
			return true
		}
	}

	return false
}

func (i *Run) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	var contextDir string
	if i.UsesBuildContext() {
		var err error
		contextDir, err = buildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return fmt.Errorf("unable to extract build context: %w", err)
		}
	}

	var addCapabilities []string
	if i.Security == dockerfile_instruction.SecurityInsecure {
		addCapabilities = []string{"all"}
	}

	if err := drv.RunCommand(ctx, containerName, i.Command, buildah.RunCommandOpts{
		CommonOpts:      drvOpts,
		ContextDir:      contextDir,
		PrependShell:    i.PrependShell,
		AddCapabilities: addCapabilities,
		NetworkType:     i.Network,
	}); err != nil {
		return fmt.Errorf("error running command %v for container %s: %w", i.Command, containerName, err)
	}

	return nil
}
