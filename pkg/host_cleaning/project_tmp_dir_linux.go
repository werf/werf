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
	"golang.org/x/sys/unix"

	"github.com/werf/werf/v3/pkg/container_backend"
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
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove project tmp entry %q: %w", path, err)
		}
		return nil
	}
	if err := checkProjectTmpMounts(path); err != nil {
		return err
	}
	if err := projectTmpRemoveAll(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrPermission) {
		return errors.Join(fmt.Errorf("remove project tmp dir %q: %w", path, err), restoreProjectTmpAge(path, info))
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
	if err := restoreProjectTmpAge(canonical, info); err != nil {
		return err
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
			return errors.Join(fmt.Errorf("remove project tmp dir %q with backend: %w", canonical, err), restoreProjectTmpAge(canonical, info))
		}
	}
	if err := os.Remove(canonical); err != nil {
		return errors.Join(fmt.Errorf("remove emptied project tmp dir %q: %w", canonical, err), restoreProjectTmpAge(canonical, info))
	}
	return nil
}

func restoreProjectTmpAge(path string, info os.FileInfo) error {
	file, err := os.OpenFile(path, unix.O_PATH|unix.O_NOFOLLOW|unix.O_DIRECTORY, 0)
	if err != nil {
		return fmt.Errorf("open project tmp dir %q to restore age: %w", path, err)
	}
	currentInfo, err := file.Stat()
	if err != nil {
		return errors.Join(fmt.Errorf("stat project tmp dir %q to restore age: %w", path, err), file.Close())
	}
	if !os.SameFile(info, currentInfo) {
		return errors.Join(fmt.Errorf("project tmp dir %q changed before restoring age", path), file.Close())
	}
	// Use the opened inode even when an unsafe ancestor is replaced.
	err = os.Chtimes(fmt.Sprintf("/proc/self/fd/%d", file.Fd()), info.ModTime(), info.ModTime())
	if err != nil {
		err = fmt.Errorf("restore project tmp dir %q age: %w", path, err)
	}
	return errors.Join(err, file.Close())
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
