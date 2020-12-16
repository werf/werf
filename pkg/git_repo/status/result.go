package status

import (
	"context"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"

	"github.com/werf/werf/pkg/path_matcher"
)

type Result struct {
	repository             *git.Repository
	repositoryAbsFilepath  string // absolute path
	repositoryFullFilepath string // path relative to main repository
	fileStatusList         git.Status
	submoduleResults       []*SubmoduleResult
}

type SubmoduleResult struct {
	*Result
	isNotInitialized bool
	isNotClean       bool
	currentCommit    string
}

type FilterOptions struct {
	StagingOnly  bool
	WorktreeOnly bool
}

func (r *Result) Status(ctx context.Context, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := &Result{
		repository:             r.repository,
		repositoryAbsFilepath:  r.repositoryAbsFilepath,
		repositoryFullFilepath: r.repositoryFullFilepath,
		fileStatusList:         git.Status{},
		submoduleResults:       []*SubmoduleResult{},
	}

	for fileStatusPath, fileStatus := range r.fileStatusList {
		fileStatusFilepath := filepath.FromSlash(fileStatusPath)
		fileStatusFullFilepath := filepath.Join(r.repositoryFullFilepath, fileStatusFilepath)

		if pathMatcher.MatchPath(fileStatusFullFilepath) {
			res.fileStatusList[fileStatusPath] = fileStatus

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

	for _, submoduleResult := range r.submoduleResults {
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleResult.repositoryFullFilepath)
		if isMatched || shouldGoThrough {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogF("Submodule was checking: %s\n", submoduleResult.repositoryFullFilepath)
			}

			if submoduleResult.isNotInitialized {
				res.submoduleResults = append(res.submoduleResults, submoduleResult)

				if debugProcess() {
					logboek.Context(ctx).Debug().LogFWithCustomStyle(
						style.Get(style.FailName),
						"Submodule is not initialized: path %s will be added to checksum\n",
						submoduleResult.repositoryFullFilepath,
					)
				}
				continue
			}

			if submoduleResult.isNotClean {
				if debugProcess() {
					logboek.Context(ctx).Debug().LogFWithCustomStyle(
						style.Get(style.FailName),
						"Submodule is not clean: current commit %s will be added to checksum\n",
						submoduleResult.currentCommit,
					)
				}
			}

			newResult, err := submoduleResult.Status(ctx, pathMatcher)
			if err != nil {
				return nil, err
			}

			newSubmoduleResult := &SubmoduleResult{
				Result:           newResult,
				isNotInitialized: false,
				isNotClean:       submoduleResult.isNotClean,
				currentCommit:    submoduleResult.currentCommit,
			}

			if !newSubmoduleResult.isEmpty(FilterOptions{}) {
				res.submoduleResults = append(res.submoduleResults, newSubmoduleResult)
			}
		}
	}

	return res, nil
}

// FilePathList method returns file paths relative to the main repository
func (r *Result) FilePathList(options FilterOptions) []string {
	var result []string
	for _, filePath := range r.filteredFilePathList(options) {
		result = append(result, filepath.Join(r.repositoryFullFilepath, filePath))
	}

	for _, submoduleResult := range r.submoduleResults {
		result = append(result, submoduleResult.FilePathList(options)...)
	}

	return result
}

// filteredFilePathList method returns file paths relative to the repository except submodules
func (r *Result) filteredFilePathList(options FilterOptions) []string {
	var result []string
	for fileStatusPath, fileStatus := range r.fileStatusList {
		if isFileStatusAccepted(fileStatus, options) {
			result = append(result, fileStatusPath)
		}
	}

	return result
}

func (r *Result) IsEmpty(options FilterOptions) bool {
	return len(r.filteredFilePathList(options)) == 0 && func() bool {
		for _, sr := range r.submoduleResults {
			if !sr.IsEmpty(options) {
				return false
			}
		}

		return true
	}()
}

func (sr *SubmoduleResult) isEmpty(options FilterOptions) bool {
	return sr.Result.IsEmpty(options) && !sr.isNotClean && !sr.isNotInitialized
}

func isFileStatusAccepted(fileStatus *git.FileStatus, options FilterOptions) bool {
	if (options.StagingOnly && !isFileStatusCodeExpected(fileStatus.Staging)) || (options.WorktreeOnly && !isFileStatusCodeExpected(fileStatus.Worktree)) {
		return false
	}

	return true
}

func isFileStatusCodeExpected(code git.StatusCode) bool {
	return !(code == git.Unmodified || code == git.Untracked)
}
