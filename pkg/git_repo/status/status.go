package status

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"

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

func Status(ctx context.Context, repository *git.Repository, repositoryAbsFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	return status(ctx, repository, repositoryAbsFilepath, "", pathMatcher)
}

func status(ctx context.Context, repository *git.Repository, repositoryAbsFilepath string, repositoryFullFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
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

	result := &Result{
		repository:             repository,
		repositoryAbsFilepath:  repositoryAbsFilepath,
		repositoryFullFilepath: repositoryFullFilepath,
		fileStatusList:         git.Status{},
		submoduleResults:       []*SubmoduleResult{},
	}

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

		if pathMatcher.MatchPath(fileStatusFullFilepath) {
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
		submoduleRepositoryAbsFilepath := filepath.Join(repositoryAbsFilepath, submoduleFilepath)

		matched, shouldGoTrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFullFilepath)
		if matched || shouldGoTrough {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogF("Submodule was checking: %s\n", submoduleFullFilepath)
			}

			submoduleResult := &SubmoduleResult{}
			submoduleRepository, err := submodule.Repository()
			if err != nil {
				if err == git.ErrSubmoduleNotInitialized {
					if debugProcess() {
						logboek.Context(ctx).Debug().LogFWithCustomStyle(
							style.Get(style.FailName),
							"Submodule is not initialized: path %s will be added to checksum\n",
							submoduleFullFilepath,
						)
					}

					submoduleResult.isNotInitialized = true
					submoduleResult.Result = &Result{
						repositoryAbsFilepath:  submoduleRepositoryAbsFilepath,
						repositoryFullFilepath: submoduleFullFilepath,
					}

					result.submoduleResults = append(result.submoduleResults, submoduleResult)
					continue
				}

				return nil, fmt.Errorf("getting submodule repository failed (%s): %s", submoduleFullFilepath, err)
			}

			submoduleStatus, err := submodule.Status()
			if err != nil {
				return nil, err
			}

			if !submoduleStatus.IsClean() {
				submoduleResult.isNotClean = true
				submoduleResult.currentCommit = submoduleStatus.Current.String()

				if debugProcess() {
					logboek.Context(ctx).Debug().LogFWithCustomStyle(
						style.Get(style.FailName),
						"Submodule is not clean: current commit %s will be added to checksum\n",
						submoduleStatus.Current,
					)
				}
			}

			sResult, err := status(ctx, submoduleRepository, submoduleRepositoryAbsFilepath, submoduleFullFilepath, pathMatcher)
			if err != nil {
				return nil, err
			}

			submoduleResult.Result = sResult

			if !submoduleResult.isEmpty(FilterOptions{}) {
				result.submoduleResults = append(result.submoduleResults, submoduleResult)
			}
		}
	}

	return result, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_STATUS_PROCESS") == "1"
}
