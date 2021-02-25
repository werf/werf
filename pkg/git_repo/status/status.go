package status

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
)

var fileStatusMapping = map[git.StatusCode]string{
	git.Unmodified:         "Unmodified",
	git.Untracked:          "Untracked",
	git.Modified:           "Modified",
	git.Added:              "Added",
	git.Deleted:            "Deleted",
	git.Renamed:            "Renamed",
	git.Copied:             "Copied",
	git.UpdatedButUnmerged: "Updated",
}

func Status(ctx context.Context, repository *git.Repository, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	return status(ctx, repository, "", pathMatcher)
}

func status(ctx context.Context, repository *git.Repository, repositoryFullFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, err
	}

	submoduleByPath := map[string]*git.Submodule{}
	for _, submodule := range submodules {
		submoduleByPath[submodule.Config().Path] = submodule
	}

	result := NewResult(repositoryFullFilepath, git.Status{}, []*SubmoduleResult{})

	worktreeStatus, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	var worktreeStatusPaths []string
	for fileStatusPath := range worktreeStatus {
		worktreeStatusPaths = append(worktreeStatusPaths, fileStatusPath)
	}

	sort.Strings(worktreeStatusPaths)

	for _, fileStatusPath := range worktreeStatusPaths {
		if _, ok := submoduleByPath[fileStatusPath]; ok {
			continue
		}

		fileStatus := worktreeStatus[fileStatusPath]
		fileStatusFilepath := filepath.FromSlash(fileStatusPath)
		fileStatusFullFilepath := filepath.Join(repositoryFullFilepath, fileStatusFilepath)

		if pathMatcher.IsPathMatched(fileStatusFullFilepath) {
			result.fileStatusList[fileStatusPath] = fileStatus

			if debugProcess() {
				logboek.Context(ctx).Debug().LogF(
					"File was added:         %s (worktree: %s, staging: %s)\n",
					fileStatusFullFilepath,
					fileStatusMapping[fileStatus.Worktree],
					fileStatusMapping[fileStatus.Staging],
				)
			}
		}
	}

	for submodulePath, submodule := range submoduleByPath {
		submoduleFilepath := filepath.FromSlash(submodulePath)
		submoduleFullFilepath := filepath.Join(repositoryFullFilepath, submoduleFilepath)

		if pathMatcher.IsDirOrSubmodulePathMatched(submoduleFullFilepath) {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogF("Submodule was checking: %s\n", submoduleFullFilepath)
			}

			submoduleRepository, err := submodule.Repository()
			if err != nil {
				return nil, fmt.Errorf("getting submodule repository failed (%s): %s", submoduleFullFilepath, err)
			}

			submoduleStatus, err := submodule.Status()
			if err != nil {
				return nil, err
			}

			if !submoduleStatus.IsClean() {
				if debugProcess() {
					logboek.Context(ctx).Debug().LogF(
						"Submodule is not clean: current commit %q expected %q\n",
						submoduleStatus.Current,
						submoduleStatus.Expected,
					)
				}
			}

			newResult, err := status(ctx, submoduleRepository, submoduleFullFilepath, pathMatcher)
			if err != nil {
				return nil, err
			}

			submoduleResult := NewSubmoduleResult(submodule.Config().Name, submodule.Config().Path, submoduleStatus, newResult)

			result.submoduleResults = append(result.submoduleResults, submoduleResult)
		}
	}

	return result, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_STATUS_PROCESS") == "1"
}
