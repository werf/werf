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

	canonical, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve project tmp dir %q: %w", path, err)
	}
	canonical, err = filepath.EvalSymlinks(canonical)
	if err != nil {
		return fmt.Errorf("resolve project tmp dir %q: %w", path, err)
	}
	currentInfo, err := os.Lstat(canonical)
	if err != nil {
		return fmt.Errorf("stat project tmp dir %q before removal: %w", canonical, err)
	}
	if !currentInfo.IsDir() || !os.SameFile(info, currentInfo) {
		return fmt.Errorf("project tmp dir %q changed before removal", path)
	}
	if err := checkProjectTmpParents(canonical); err != nil {
		return err
	}
	if err := checkProjectTmpMounts(canonical); err != nil {
		return err
	}
	entries, err := os.ReadDir(canonical)
	if err != nil {
		return fmt.Errorf("read project tmp dir %q: %w", canonical, err)
	}
	dirs := lo.Map(entries, func(entry os.DirEntry, _ int) string {
		return filepath.Join(canonical, entry.Name())
	})
	if len(dirs) > 0 {
		if err := backend.RemoveHostDirs(ctx, canonical, dirs); err != nil {
			return errors.Join(fmt.Errorf("remove project tmp dir %q with backend: %w", canonical, err), os.Chtimes(canonical, info.ModTime(), info.ModTime()))
		}
	}
	if err := os.Remove(canonical); err != nil {
		return errors.Join(fmt.Errorf("remove emptied project tmp dir %q: %w", canonical, err), os.Chtimes(canonical, info.ModTime(), info.ModTime()))
	}
	return nil
}

func checkProjectTmpParents(path string) error {
	// A writable non-sticky ancestor lets another user replace the bind source.
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		info, err := projectTmpLstat(parent)
		if err != nil {
			return fmt.Errorf("stat project tmp parent %q: %w", parent, err)
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !info.IsDir() || !ok || (stat.Uid != 0 && stat.Uid != uint32(os.Geteuid())) {
			return fmt.Errorf("preserve project tmp dir %q: parent %q is not owned by root or the current user", path, parent)
		}
		if info.Mode().Perm()&0o022 != 0 && info.Mode()&os.ModeSticky == 0 {
			return fmt.Errorf("preserve project tmp dir %q: parent %q is writable without the sticky bit", path, parent)
		}
		if parent == filepath.Dir(parent) {
			return nil
		}
	}
}

func checkProjectTmpMounts(path string) error {
	realPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve project tmp dir %q: %w", path, err)
	}
	realPath, err = filepath.EvalSymlinks(realPath)
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
