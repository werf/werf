package status

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/path_matcher"
)

var fileStatusMapping = map[rune]string{
	' ': "Unmodified",
	'?': "Untracked",
	'M': "Modified",
	'A': "Added",
	'D': "Deleted",
	'R': "Renamed",
	'C': "Copied",
	'U': "Updated",
}

func Status(repository *git.Repository, repositoryAbsFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	return status(repository, repositoryAbsFilepath, "", pathMatcher)
}

func status(repository *git.Repository, repositoryAbsFilepath string, repositoryFullFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, err
	}

	submoduleList := map[string]*git.Submodule{}
	for _, submodule := range submodules {
		submoduleList[submodule.Config().Path] = submodule
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
		if _, ok := submoduleList[fileStatusPath]; ok {
			continue
		}

		fileStatus := worktreeStatus[fileStatusPath]
		fileStatusFilepath := filepath.FromSlash(fileStatusPath)
		fileStatusFullFilepath := filepath.Join(repositoryFullFilepath, fileStatusFilepath)

		if pathMatcher.MatchPath(fileStatusFullFilepath) {
			result.fileStatusList[fileStatusPath] = fileStatus

			if debugProcess() {
				logboek.Debug.LogF(
					"File was added:         %s (worktree: %s, staging: %s)\n",
					fileStatusFullFilepath,
					fileStatusMapping[rune(fileStatus.Worktree)],
					fileStatusMapping[rune(fileStatus.Staging)],
				)
			}
		}
	}

	for submodulePath, submodule := range submoduleList {
		submoduleFilepath := filepath.FromSlash(submodulePath)
		submoduleFullFilepath := filepath.Join(repositoryFullFilepath, submoduleFilepath)
		submoduleRepositoryAbsFilepath := filepath.Join(repositoryAbsFilepath, submoduleFilepath)

		matched, shouldGoTrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFullFilepath)
		if matched || shouldGoTrough {
			if debugProcess() {
				logboek.Debug.LogF("Submodule was checking: %s\n", submoduleFullFilepath)
			}

			submoduleResult := &SubmoduleResult{}
			submoduleRepository, err := submodule.Repository()
			if err != nil {
				if err == git.ErrSubmoduleNotInitialized {
					if debugProcess() {
						logboek.Debug.LogFWithCustomStyle(
							logboek.StyleByName(logboek.FailStyleName),
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
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not clean: current commit %s will be added to checksum\n",
						submoduleStatus.Current,
					)
				}
			}

			sResult, err := status(submoduleRepository, submoduleRepositoryAbsFilepath, submoduleFullFilepath, pathMatcher)
			if err != nil {
				return nil, err
			}

			submoduleResult.Result = sResult

			if !submoduleResult.isEmpty() {
				result.submoduleResults = append(result.submoduleResults, submoduleResult)
			}
		}
	}

	return result, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_STATUS_PROCESS") == "1"
}
