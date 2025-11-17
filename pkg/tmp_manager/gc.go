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
)

var (
	ErrPathRemoval = errors.New("path removal")

	timeSince = time.Since // for stubbing in tests
)

func ShouldRunAutoGC() (bool, error) {
	projectDirsToRemove, pathsToRemove, err := collectPaths()
	if err != nil {
		return false, fmt.Errorf("collect paths: %w", err)
	}
	return len(projectDirsToRemove) > 0 || len(pathsToRemove) > 0, nil
}

func RunGC(ctx context.Context, dryRun bool) error {
	projectDirsToRemove, pathsToRemove, err := collectPaths()
	if err != nil {
		return fmt.Errorf("collect paths: %w", err)
	}

	return runGCForPaths(ctx, dryRun, slices.Concat(projectDirsToRemove, pathsToRemove))
}

func runGCForPaths(ctx context.Context, dryRun bool, paths []string) error {
	removeErrors := make([]error, 0, len(paths))

	for _, path := range paths {
		logboek.Context(ctx).Default().LogLn(path)

		if dryRun {
			continue
		}

		if err := os.RemoveAll(path); err != nil {
			removeErrors = append(removeErrors, errors.Join(ErrPathRemoval, err))
		}
	}

	return errors.Join(removeErrors...) // magic of errors.Join(): omit nil errors if they exist
}

func collectPaths() ([]string, []string, error) {
	gcPathList := []gcPath{
		newGCPath(filepath.Join(getReleasedTmpDirs(), projectsServiceDir), 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), projectsServiceDir), 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), dockerConfigsServiceDir), time.Hour*4),
		newGCPath(filepath.Join(getCreatedTmpDirs(), kubeConfigsServiceDir), 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), werfConfigRendersServiceDir), 0),
		newGCPath(filepath.Join(getCreatedTmpDirs(), contextArchivesDir), 0),
		newGCPath(getContextTmpDir(), 0), // TODO: backward compatible cleaning (will be dropped in v3)
	}

	dirSlices := make([][]string, 0, len(gcPathList))
	symlinkSlices := make([][]string, 0, len(gcPathList))

	for _, gcPathItem := range gcPathList {
		dirs, symlinks, err := listDirAndFollowSymlinks(gcPathItem.path, gcPathItem.keepingTime)
		if err != nil {
			return nil, nil, fmt.Errorf("list and filter path %v: %w", gcPathItem.path, err)
		}
		dirSlices = append(dirSlices, dirs)
		symlinkSlices = append(symlinkSlices, symlinks)
	}

	return slices.Concat(dirSlices...), slices.Concat(symlinkSlices...), nil
}

// listDirAndFollowSymlinks returns list of dirs and symlinks
func listDirAndFollowSymlinks(dir string, minFileAge time.Duration) ([]string, []string, error) {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
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
		info, err := dirEntry.Info()
		if err != nil {
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
		default:
			listOfDirs = append(listOfDirs, linkOrFilePath)
			// resolve only symlinks
			continue
		}

		filePath, err := os.Readlink(linkOrFilePath)
		if err != nil {
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
	path        string
	keepingTime time.Duration
}

func newGCPath(path string, keepingTime time.Duration) gcPath {
	return gcPath{
		path:        path,
		keepingTime: keepingTime,
	}
}
