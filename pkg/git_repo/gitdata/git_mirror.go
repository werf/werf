package gitdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/volumeutils"
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
func GetGitMirrorsAndRemoveInvalid(ctx context.Context, cacheVersionRoot string) ([]GitDataEntry, error) {
	var res []GitDataEntry

	fileStat, err := os.Stat(cacheVersionRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}
	if !fileStat.IsDir() {
		logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", cacheVersionRoot)
		if err := os.RemoveAll(cacheVersionRoot); err != nil {
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
			if err := os.RemoveAll(repoPath); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
			continue
		}

		repoChildren, err := ioutil.ReadDir(repoPath)
		if err != nil {
			return nil, fmt.Errorf("error reading dir %q: %w", repoPath, err)
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
				if err := os.RemoveAll(childPath); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", childPath, err)
				}
			}
		}

		if !shallowFound && !markerFound {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: no shallow mirror and no requires_full marker inside\n", repoPath)
			if err := os.RemoveAll(repoPath); err != nil {
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
			return nil, fmt.Errorf("error getting dir %q size: %w", shallowPath, err)
		}

		lastAccessAtPath := filepath.Join(shallowPath, "last_access_at")
		lastAccessAt, err := timestamps.ReadTimestampFile(lastAccessAtPath)
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: error reading last access timestamp file %q: %v\n", shallowPath, lastAccessAtPath, err)
			if err := os.RemoveAll(shallowPath); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", shallowPath, err)
			}
			continue
		}

		res = append(res, &GitRepoDesc{
			Path:          shallowPath,
			Size:          size,
			LastAccessAt:  lastAccessAt,
			CacheBasePath: cacheVersionRoot,
		})
	}

	return res, nil
}
