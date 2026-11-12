package instruction

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/tmp_manager"
)

type Copy struct {
	instructions.CopyCommand
}

func NewCopy(i instructions.CopyCommand) *Copy {
	return &Copy{CopyCommand: i}
}

func (i *Copy) UsesBuildContext() bool {
	return i.From == "" && len(i.SourcePaths) > 0
}

func (i *Copy) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if len(i.SourcePaths) > 0 {
		if err := i.copySourcePaths(ctx, containerName, drv, drvOpts, buildContextArchive); err != nil {
			return err
		}
	}

	return copySourceContents(ctx, containerName, drv, drvOpts, i.SourceContents, i.DestPath, i.Chown, i.Chmod)
}

func (i *Copy) copySourcePaths(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
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

// copySourceContents writes heredoc sources to files the way BuildKit does (mode 0644, named after
// the heredoc delimiter) and copies them to the destination.
func copySourceContents(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, sourceContents []instructions.SourceContent, destPath, chown, chmod string) error {
	if len(sourceContents) == 0 {
		return nil
	}

	contentsDir, err := tmp_manager.TempDir("heredoc-*")
	if err != nil {
		return fmt.Errorf("create heredoc sources dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(contentsDir); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove heredoc sources dir %s: %s\n", contentsDir, err)
		}
	}()

	sourcePaths := make([]string, 0, len(sourceContents))
	for _, content := range sourceContents {
		if !filepath.IsLocal(content.Path) {
			return fmt.Errorf("invalid heredoc name %q: must be a relative path without \"..\"", content.Path)
		}

		contentPath := filepath.Join(contentsDir, content.Path)
		if err := os.MkdirAll(filepath.Dir(contentPath), os.ModePerm); err != nil {
			return fmt.Errorf("create dir for heredoc source %q: %w", content.Path, err)
		}
		if err := os.WriteFile(contentPath, []byte(content.Data), 0o644); err != nil {
			return fmt.Errorf("write heredoc source %q: %w", content.Path, err)
		}
		// buildah keeps the source mode without --chmod, so it must not depend on the umask
		if err := os.Chmod(contentPath, 0o644); err != nil {
			return fmt.Errorf("chmod heredoc source %q: %w", content.Path, err)
		}
		sourcePaths = append(sourcePaths, content.Path)
	}

	if err := drv.Copy(ctx, containerName, contentsDir, sourcePaths, destPath, buildah.CopyOpts{
		CommonOpts: drvOpts,
		Chown:      chown,
		Chmod:      chmod,
	}); err != nil {
		return fmt.Errorf("error copying heredoc sources %v to %s for container %s: %w", sourcePaths, destPath, containerName, err)
	}

	return nil
}
