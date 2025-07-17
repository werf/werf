package tmp_manager

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/werf"
)

var timeSince = time.Since // for stubbing in tests

func ShouldRunAutoGC() (bool, error) {
	projectDirsToRemove, pathsToRemove, err := collectPaths()
	if err != nil {
		return false, fmt.Errorf("collect paths: %w", err)
	}
	return len(projectDirsToRemove) > 0 || len(pathsToRemove) > 0, nil
}

func RunGC(ctx context.Context, dryRun bool, containerBackend container_backend.ContainerBackend) error {
	projectDirsToRemove, pathsToRemove, err := collectPaths()
	if err != nil {
		return fmt.Errorf("collect paths: %w", err)
	}

	for _, itemToRemove := range slices.Concat(projectDirsToRemove, pathsToRemove) {
		logboek.Context(ctx).Default().LogLn(itemToRemove)
	}

	if dryRun {
		return nil
	}

	removeErrors := make([]error, 0, len(projectDirsToRemove)+len(pathsToRemove))

	if len(projectDirsToRemove) > 0 {
		if runtime.GOOS == "windows" {
			for _, path := range projectDirsToRemove {
				if err = os.RemoveAll(path); err != nil {
					removeErrors = append(removeErrors, fmt.Errorf("unable to remove tmp project dir %s: %w", path, err))
				}
			}
		} else {
			if err := containerBackend.RemoveHostDirs(ctx, werf.GetTmpDir(), projectDirsToRemove); err != nil {
				removeErrors = append(removeErrors, fmt.Errorf("unable to remove tmp projects dirs %s: %w", strings.Join(projectDirsToRemove, ", "), err))
			}
		}
	}

	for _, path := range pathsToRemove {
		if err = os.RemoveAll(path); err != nil {
			removeErrors = append(removeErrors, fmt.Errorf("unable to remove path %s: %w", path, err))
		}
	}

	return errors.Join(removeErrors...) // magic of errors.Join(): omit nil errors if they exist
}

func collectPaths() ([]string, []string, error) {
	releasedProjects, err := listAndFilterPaths(filepath.Join(GetReleasedTmpDirs(), projectsServiceDir))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get released tmp projects dirs: %w", err)
	}
	createdProjects, err := listAndFilterPaths(filepath.Join(GetCreatedTmpDirs(), projectsServiceDir))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get created tmp projects dirs: %w", err)
	}
	createdDockerConfigs, err := listAndFilterPaths(filepath.Join(GetCreatedTmpDirs(), dockerConfigsServiceDir))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get created tmp docker configs: %w", err)
	}
	kubeConfigs, err := listAndFilterPaths(filepath.Join(GetCreatedTmpDirs(), kubeConfigsServiceDir))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get created tmp kubeconfigs: %w", err)
	}
	createdWerfConfigRenders, err := listAndFilterPaths(filepath.Join(GetCreatedTmpDirs(), werfConfigRendersServiceDir))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get created tmp werf config render files: %w", err)
	}

	pathDescs := slices.Concat(releasedProjects, createdProjects, createdDockerConfigs, kubeConfigs, createdWerfConfigRenders)

	dirs := make([]string, 0, len(pathDescs))
	files := make([]string, 0, len(pathDescs))

	for _, pd := range pathDescs {
		if pd.IsDir {
			dirs = append(dirs, pd.FullPath)
		} else {
			files = append(files, pd.FullPath)
		}
	}

	return slices.Clip(dirs), slices.Clip(files), nil
}

type PathDesc struct {
	IsDir    bool
	FullPath string
}

func listAndFilterPaths(dir string) ([]PathDesc, error) {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat %v dir: %w", dir, err)
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir in %s: %w", dir, err)
	}

	list := make([]PathDesc, 0, len(dirEntries)*2)

	for _, dirEntry := range dirEntries {
		info, err := dirEntry.Info()
		if err != nil {
			return nil, fmt.Errorf("file info for %s: %w", dirEntry.Name(), err)
		}

		// filter out recent files
		if timeSince(info.ModTime()) < 2*time.Hour {
			continue
		}

		linkOrFileDesc := PathDesc{
			IsDir:    info.IsDir(),
			FullPath: filepath.Join(dir, info.Name()),
		}
		list = append(list, linkOrFileDesc)

		// resolve only symlinks
		if info.Mode().Type() != os.ModeSymlink {
			continue
		}

		fileDesc := PathDesc{}
		if fileDesc.FullPath, err = os.Readlink(linkOrFileDesc.FullPath); err != nil {
			return nil, fmt.Errorf("read link %s: %w", linkOrFileDesc.FullPath, err)
		}
		if stat, err := os.Stat(fileDesc.FullPath); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("stat %q path: %w", fileDesc.FullPath, err)
		} else {
			fileDesc.IsDir = stat.IsDir()
		}

		list = append(list, fileDesc)
	}

	return slices.Clip(list), nil
}
