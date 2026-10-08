package host_cleaning

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/moby/sys/mountinfo"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/container_backend"
)

var (
	projectTmpMounts    = mountinfo.GetMounts
	projectTmpLstat     = os.Lstat
	projectTmpRemoveAll = os.RemoveAll
)

func removeProjectTmpDir(ctx context.Context, backend container_backend.ContainerBackend, path string) error {
	info, err := projectTmpLstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat project tmp dir %q: %w", path, err)
	}
	if !info.IsDir() {
		return os.RemoveAll(path)
	}
	if err := checkProjectTmpMounts(path); err != nil {
		return err
	}
	if err := projectTmpRemoveAll(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrPermission) {
		return err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return fmt.Errorf("remove project tmp dir %q: owner differs from current user: %w", path, fs.ErrPermission)
	}

	// A private sibling pins the bind source against replacement by another tmp user.
	stage, err := os.MkdirTemp(filepath.Dir(path), filepath.Base(path)+"-cleanup-")
	if err != nil {
		return fmt.Errorf("create project tmp cleanup dir: %w", err)
	}
	moved := filepath.Join(stage, "data")
	if err := os.Rename(path, moved); err != nil {
		return errors.Join(fmt.Errorf("stage project tmp dir %q: %w", path, err), os.Remove(stage))
	}
	if err := os.Chtimes(stage, info.ModTime(), info.ModTime()); err != nil {
		return fmt.Errorf("preserve age of staged project tmp dir %q: %w", stage, err)
	}
	movedInfo, err := os.Lstat(moved)
	if err != nil {
		return fmt.Errorf("stat staged project tmp dir %q: %w", moved, err)
	}
	if !movedInfo.IsDir() || !os.SameFile(info, movedInfo) {
		return fmt.Errorf("staged project tmp dir %q changed before removal", moved)
	}
	if err := checkProjectTmpMounts(moved); err != nil {
		return err
	}
	entries, err := os.ReadDir(moved)
	if err != nil {
		return fmt.Errorf("read staged project tmp dir %q: %w", moved, err)
	}
	dirs := lo.Map(entries, func(entry os.DirEntry, _ int) string {
		return filepath.Join(moved, entry.Name())
	})
	if len(dirs) > 0 {
		if err := backend.RemoveHostDirs(ctx, moved, dirs); err != nil {
			return fmt.Errorf("remove staged project tmp dir %q with backend: %w", moved, err)
		}
	}
	if err := os.Remove(moved); err != nil {
		return fmt.Errorf("remove emptied project tmp dir %q: %w", moved, err)
	}
	if err := os.Remove(stage); err != nil {
		return fmt.Errorf("remove project tmp cleanup dir %q: %w", stage, err)
	}
	return nil
}

func checkProjectTmpMounts(path string) error {
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("resolve project tmp dir %q: %w", path, err)
	}
	mounts, err := projectTmpMounts(mountinfo.PrefixFilter(realPath))
	if err != nil {
		return fmt.Errorf("check project tmp mounts %q: %w", path, err)
	}
	if len(mounts) > 0 {
		return fmt.Errorf("preserve project tmp dir %q containing mountpoint %q", path, mounts[0].Mountpoint)
	}
	return nil
}
