package tmp_manager

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf"
)

var (
	ErrPathRemoval = errors.New("path removal")

	timeSince    = time.Since // for stubbing in tests
	onSameDevice = sameDevice
)

func ShouldRunAutoGC() (bool, error) {
	projectDirsToRemove, pathsToRemove, err := collectPaths()
	if err != nil {
		return false, fmt.Errorf("collect paths: %w", err)
	}
	return len(projectDirsToRemove) > 0 || len(pathsToRemove) > 0, nil
}

type RunGCOptions struct {
	DryRun           bool
	RemoveProjectDir func(context.Context, string) error
}

func RunGC(ctx context.Context, options RunGCOptions) error {
	projectDirsToRemove, pathsToRemove, err := collectPaths()
	if err != nil {
		return fmt.Errorf("collect paths: %w", err)
	}

	return runGCForPaths(ctx, options, slices.Concat(projectDirsToRemove, pathsToRemove))
}

func runGCForPaths(ctx context.Context, options RunGCOptions, paths []string) error {
	removeErrors := make([]error, 0, len(paths))

	for _, path := range paths {
		logboek.Context(ctx).Default().LogLn(path)

		if options.DryRun {
			continue
		}

		remove := func() error { return os.RemoveAll(path) }
		if options.RemoveProjectDir != nil && isProjectTmpDir(path) {
			remove = func() error { return options.RemoveProjectDir(ctx, path) }
		}
		if err := remove(); err != nil {
			removeErrors = append(removeErrors, errors.Join(ErrPathRemoval, err))
		}
	}

	return errors.Join(removeErrors...) // magic of errors.Join(): omit nil errors if they exist
}

func isProjectTmpDir(path string) bool {
	for _, pattern := range []string{"werf-*-project-data-*", "werf-project-data-*"} {
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err == nil && matched {
			return true
		}
	}
	return false
}

func collectPaths() ([]string, []string, error) {
	gcPathList := []gcPath{
		newGCPath(filepath.Join(getReleasedTmpDirs(), projectsServiceDir), "", 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), projectsServiceDir), "", 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), dockerConfigsServiceDir), "", time.Hour*6),
		newGCPath(filepath.Join(getCreatedTmpDirs(), kubeConfigsServiceDir), "", 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), werfConfigRendersServiceDir), "", 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), contextArchivesDir), "", 0),
		newGCPath(getContextTmpDir(), "", 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), contextPinsServiceDir), "", 0),
		newGCPath(filepath.Join(getServiceTmpDir(), contextPinsServiceDir), "", contextPinMaxAge),
		// Project dirs are not registered either until the command delegates the cleanup, and they
		// hold the pinned git inputs of a build. They live directly in the tmp dir shared with
		// everything else on the host, so only our own prefix is swept and no symlink is followed.
		newNoFollowGCPath(werf.GetTmpDir(), "werf-*-project-data-*", projectDirMaxAge),
		newNoFollowGCPath(werf.GetTmpDir(), "werf-project-data-*", projectDirMaxAge),
	}

	dirSlices := make([][]string, 0, len(gcPathList))
	symlinkSlices := make([][]string, 0, len(gcPathList))

	for _, gcPathItem := range gcPathList {
		dirs, symlinks, err := listDirAndFollowSymlinks(gcPathItem)
		if err != nil {
			return nil, nil, fmt.Errorf("list and filter path %v: %w", gcPathItem.path, err)
		}
		dirSlices = append(dirSlices, dirs)
		symlinkSlices = append(symlinkSlices, symlinks)
	}

	return slices.Concat(dirSlices...), slices.Concat(symlinkSlices...), nil
}

// listDirAndFollowSymlinks returns list of dirs and symlinks. With a non-empty namePattern only
// matching entries are collected, which is what makes a dir shared with foreign files sweepable.
// Symlink targets are collected only for registry dirs, where werf itself wrote the links; sweeping
// a dir werf does not own must never delete whatever a foreign link happens to point at.
func listDirAndFollowSymlinks(gcPathItem gcPath) ([]string, []string, error) {
	dir, namePattern, minFileAge := gcPathItem.path, gcPathItem.namePattern, gcPathItem.keepingTime

	dirInfo, err := os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, fmt.Errorf("stat %v dir: %w", dir, err)
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read dir in %s: %w", dir, err)
	}

	listOfDirs := make([]string, 0, len(dirEntries))
	listOfSymlinks := make([]string, 0, len(dirEntries))

	for _, dirEntry := range dirEntries {
		if namePattern != "" {
			if matched, err := filepath.Match(namePattern, dirEntry.Name()); err != nil {
				return nil, nil, fmt.Errorf("match GC path pattern %q: %w", namePattern, err)
			} else if !matched {
				continue
			}
		}

		info, err := dirEntry.Info()
		if errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, nil, fmt.Errorf("file info for %s: %w", dirEntry.Name(), err)
		}

		// filter out recent files
		if timeSince(info.ModTime()) < minFileAge {
			continue
		}

		linkOrFilePath := filepath.Join(dir, dirEntry.Name())

		switch info.Mode().Type() {
		case os.ModeSymlink:
			listOfSymlinks = append(listOfSymlinks, linkOrFilePath)
			if !gcPathItem.followSymlinks {
				continue
			}
		default:
			// A filesystem mounted under a dir werf does not own keeps its data elsewhere; a
			// recursive removal would empty it instead of reclaiming tmp space.
			if !gcPathItem.followSymlinks && !onSameDevice(dirInfo, info) {
				continue
			}
			listOfDirs = append(listOfDirs, linkOrFilePath)
			continue
		}

		filePath, err := os.Readlink(linkOrFilePath)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, nil, fmt.Errorf("read link %s: %w", linkOrFilePath, err)
		}
		if _, err = os.Stat(filePath); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, nil, fmt.Errorf("stat %v dir: %w", filePath, err)
		}

		listOfDirs = append(listOfDirs, filePath)
	}

	return slices.Clip(listOfDirs), slices.Clip(listOfSymlinks), nil
}

type gcPath struct {
	path           string
	namePattern    string
	keepingTime    time.Duration
	followSymlinks bool
}

func newGCPath(path, namePattern string, keepingTime time.Duration) gcPath {
	return gcPath{
		path:           path,
		namePattern:    namePattern,
		keepingTime:    keepingTime,
		followSymlinks: true,
	}
}

func newNoFollowGCPath(path, namePattern string, keepingTime time.Duration) gcPath {
	return gcPath{
		path:        path,
		namePattern: namePattern,
		keepingTime: keepingTime,
	}
}
