package status

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/src-d/go-git.v4"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/path_matcher"
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

func Status(repository *git.Repository, repositoryFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	return status(repository, repositoryFilepath, "", pathMatcher)
}

func status(repository *git.Repository, repositoryFilepath string, relToBaseRepositoryFilepath string, pathMatcher path_matcher.PathMatcher) (*Result, error) {
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
		repository:         repository,
		repositoryFilepath: repositoryFilepath,
		fileStatusList:     git.Status{},
		submoduleResults:   []*SubmoduleResult{},
	}

	worktreeStatus, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	var worktreeStatusPaths []string
	for path, _ := range worktreeStatus {
		worktreeStatusPaths = append(worktreeStatusPaths, path)
	}

	sort.Strings(worktreeStatusPaths)

	for _, path := range worktreeStatusPaths {
		if _, ok := submoduleList[path]; ok {
			continue
		}

		fileStatus := worktreeStatus[path]

		// r prefix == relative to base repo path
		rFilepath := filepath.Join(relToBaseRepositoryFilepath, filepath.FromSlash(path))
		if pathMatcher.MatchPath(rFilepath) {
			result.fileStatusList[rFilepath] = fileStatus

			if debugProcess() {
				logboek.Debug.LogF("File was added:         %s (worktree: %s, staging: %s)\n", rFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)])
			}
		}
	}

	for path, submodule := range submoduleList {
		// r prefix == relative to base repo path
		rSubmoduleFilepath := filepath.Join(relToBaseRepositoryFilepath, filepath.FromSlash(path))
		matched, shouldGoTrough := pathMatcher.ProcessDirOrSubmodulePath(rSubmoduleFilepath)
		if matched || shouldGoTrough {
			if debugProcess() {
				logboek.Debug.LogF("Submodule was checking: %s\n", rSubmoduleFilepath)
			}

			sSubmoduleResult := &SubmoduleResult{
				relToBaseRepositorySubmoduleFilepath: rSubmoduleFilepath,
			}

			submoduleRepository, err := submodule.Repository()
			if err != nil {
				if err == git.ErrSubmoduleNotInitialized {
					if debugProcess() {
						logboek.Debug.LogFWithCustomStyle(
							logboek.StyleByName(logboek.FailStyleName),
							"Submodule is not initialized: path %s will be added to checksum\n",
							rSubmoduleFilepath,
						)
					}

					sSubmoduleResult.isNotInitialized = true
					result.submoduleResults = append(result.submoduleResults, sSubmoduleResult)
					continue
				}
				return nil, err
			}

			submoduleStatus, err := submodule.Status()
			if err != nil {
				return nil, err
			}

			if !submoduleStatus.IsClean() {
				sSubmoduleResult.isNotClean = true
				sSubmoduleResult.currentCommit = submoduleStatus.Current.String()

				if debugProcess() {
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not clean: current commit %s will be added to checksum\n",
						submoduleStatus.Current,
					)
				}
			}

			submoduleRepositoryFilepath := filepath.Join(repositoryFilepath, filepath.FromSlash(path))
			sResult, err := status(submoduleRepository, submoduleRepositoryFilepath, rSubmoduleFilepath, pathMatcher)
			if err != nil {
				return nil, err
			}

			sSubmoduleResult.Result = sResult

			if !sSubmoduleResult.isEmpty() {
				result.submoduleResults = append(result.submoduleResults, sSubmoduleResult)
			}
		}
	}

	return result, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_STATUS_PROCESS") == "1"
}
