package instruction

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/container_backend"
)

type Copy struct {
	instructions.CopyCommand
}

func NewCopy(i instructions.CopyCommand) *Copy {
	return &Copy{CopyCommand: i}
}

func (i *Copy) UsesBuildContext() bool {
	return i.From == ""
}

func (i *Copy) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	var contextDir string
	var sourceContainer string
	if i.UsesBuildContext() {
		var err error
		contextDir, err = buildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return fmt.Errorf("unable to extract build context: %w", err)
		}
	} else {
		container, err := drv.FromCommand(ctx, "", i.From, buildah.FromCommandOpts(drvOpts))
		if err != nil {
			return fmt.Errorf("unable to create container from image %q: %w", i.From, err)
		}
		sourceContainer = container

		contextDir, err = drv.Mount(ctx, sourceContainer, buildah.MountOpts(drvOpts))
		if err != nil {
			rmErr := drv.Rm(ctx, sourceContainer, buildah.RmOpts(drvOpts))
			if rmErr != nil {
				return errors.Join(
					fmt.Errorf("unable to mount container %q: %w", sourceContainer, err),
					fmt.Errorf("unable to remove container %q: %w", sourceContainer, rmErr),
				)
			}

			return fmt.Errorf("unable to mount container %q: %w", sourceContainer, err)
		}
	}

	copyErr := drv.Copy(ctx, containerName, contextDir, i.SourcePaths, i.DestPath, buildah.CopyOpts{
		CommonOpts: drvOpts,
		Chown:      i.Chown,
		Chmod:      i.Chmod,
	})

	var cleanupErr error
	if sourceContainer != "" {
		if err := drv.Umount(ctx, sourceContainer, buildah.UmountOpts(drvOpts)); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("unable to unmount container %q: %w", sourceContainer, err))
		}

		if err := drv.Rm(ctx, sourceContainer, buildah.RmOpts(drvOpts)); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("unable to remove container %q: %w", sourceContainer, err))
		}
	}

	if copyErr != nil {
		copyErr = fmt.Errorf("error copying %v to %s for container %s: %w", i.SourcePaths, i.DestPath, containerName, copyErr)
		if cleanupErr != nil {
			return errors.Join(copyErr, cleanupErr)
		}

		return copyErr
	}

	if cleanupErr != nil {
		return cleanupErr
	}

	return nil
}
