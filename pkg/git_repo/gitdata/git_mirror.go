package gitdata

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/volumeutils"
)

// GetGitMirrorsAndRemoveInvalid scans the given cacheVersionRoot directory and
// returns a list of GitRepoDesc for each valid shallow git mirror found. It
// removes invalid entries and handles errors appropriately.
//
// The directory structure expected is as follows:
// ├── c447df0d5918decb5d832cb4324e3e2cbe0670eb3fe9301f795be831a9175f47
// │   ├── shallow/
// │   │   └── ... (repository files)
// │   └── requires_full (optional marker file)
// └── ... (other repositories)
//
// Each shallow mirror is an LRU entry with its own last_access_at. The
// requires_full marker is persistent metadata, not an LRU entry: a repo dir
// holding only the marker is valid and kept. A repo dir with neither shallow
// mirror nor marker is removed.
//
// An entry whose size or access marker cannot be read is preserved and left
// out of the result; its error is joined into the returned error.
func GetGitMirrorsAndRemoveInvalid(ctx context.Context, cacheVersionRoot string, options ScanOptions) ([]GitDataEntry, error) {
	var res []GitDataEntry
	var errs []error

	fileStat, err := os.Stat(cacheVersionRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}
	if !fileStat.IsDir() {
		logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", cacheVersionRoot)
		if err := removePath(cacheVersionRoot, options); err != nil {
			return nil, fmt.Errorf("unable to remove %q: %w", cacheVersionRoot, err)
		}
		return nil, nil
	}

	repoDirs, err := ioutil.ReadDir(cacheVersionRoot)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheVersionRoot, err)
	}

	for _, repoDirInfo := range repoDirs {
		repoPath := filepath.Join(cacheVersionRoot, repoDirInfo.Name())

		if !repoDirInfo.IsDir() {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", repoPath)
			if err := removePath(repoPath, options); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
			continue
		}

		repoChildren, err := ioutil.ReadDir(repoPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("read dir %q: %w", repoPath, err))
			continue
		}

		var shallowFound, markerFound bool

		for _, childInfo := range repoChildren {
			childPath := filepath.Join(repoPath, childInfo.Name())

			switch {
			case childInfo.Name() == "shallow" && childInfo.IsDir():
				shallowFound = true
			case childInfo.Name() == "requires_full" && childInfo.Mode().IsRegular():
				markerFound = true
			default:
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q\n", childPath)
				if err := removePath(childPath, options); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", childPath, err)
				}
			}
		}

		if !shallowFound && !markerFound {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: no shallow mirror and no requires_full marker inside\n", repoPath)
			if err := removePath(repoPath, options); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
			continue
		}

		if !shallowFound {
			continue
		}

		shallowPath := filepath.Join(repoPath, "shallow")

		size, err := volumeutils.DirSizeBytes(shallowPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("get dir %q size: %w", shallowPath, err))
			continue
		}

		lastAccessAtPath := filepath.Join(shallowPath, "last_access_at")
		lastAccessAt, err := readLastAccessAt(lastAccessAtPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("read repository access timestamp %q: %w", lastAccessAtPath, err))
			continue
		}

		res = append(res, &GitRepoDesc{
			Path:          shallowPath,
			Size:          size,
			LastAccessAt:  lastAccessAt,
			CacheBasePath: cacheVersionRoot,
		})
	}

	return res, errors.Join(errs...)
}
