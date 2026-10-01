package instruction

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/container_backend"
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
	if i.UsesBuildContext() {
		var err error
		contextDir, err = buildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return fmt.Errorf("unable to extract build context: %w", err)
		}
	} else {
		container, err := drv.FromCommand(ctx, "", i.From, buildah.FromCommandOpts{})
		mounted := false
		if container != "" {
			defer func() {
				if cleanupErr := cleanupCopySource(ctx, drv, container, mounted); cleanupErr != nil {
					logboek.Context(ctx).Error().LogF("ERROR: cleanup COPY --from=%q source container %q: %s\n", i.From, container, cleanupErr)
				}
			}()
		}
		if err != nil {
			return fmt.Errorf("unable to create container from image %q: %w", i.From, err)
		}

		contextDir, err = drv.Mount(ctx, container, buildah.MountOpts{})
		if err != nil {
			return fmt.Errorf("unable to mount container %q: %w", container, err)
		}
		mounted = true
	}

	copyErr := drv.Copy(ctx, containerName, contextDir, i.SourcePaths, i.DestPath, buildah.CopyOpts{
		CommonOpts: drvOpts,
		Chown:      i.Chown,
		Chmod:      i.Chmod,
		Parents:    i.Parents,
		Ignores:    contextRelativeExcludes(i.SourcePaths, i.ExcludePatterns),
	})
	if copyErr != nil {
		copyErr = fmt.Errorf("error copying %v to %s for container %s: %w", i.SourcePaths, i.DestPath, containerName, copyErr)
	}

	return copyErr
}

func cleanupCopySource(ctx context.Context, drv buildah.Buildah, container string, mounted bool) error {
	cleanupCtx := context.WithoutCancel(ctx)
	var cleanupErrors []error
	if mounted {
		if err := drv.Umount(cleanupCtx, container, buildah.UmountOpts{}); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("unmount copy source container %q: %w", container, err))
		}
	}
	if err := drv.Rm(cleanupCtx, container, buildah.RmOpts{}); err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("remove copy source container %q: %w", container, err))
	}
	return errors.Join(cleanupErrors...)
}
