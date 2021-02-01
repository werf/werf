package status

import (
	"context"
	"path"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
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
	*git.SubmoduleStatus
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

			newResult, err := submoduleResult.Status(ctx, pathMatcher)
			if err != nil {
				return nil, err
			}

			newSubmoduleResult := &SubmoduleResult{
				Result:          newResult,
				SubmoduleStatus: submoduleResult.SubmoduleStatus,
			}

			res.submoduleResults = append(res.submoduleResults, newSubmoduleResult)
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

func (r *Result) CheckIfFilePathInsideSubmodule(relPath string) (bool, bool, string) {
	for _, sr := range r.submoduleResults {
		if util.IsSubpathOfBasePath(filepath.ToSlash(sr.repositoryFullFilepath), filepath.ToSlash(relPath)) {
			submodulePath := sr.repositoryFullFilepath

			if sr.Current != sr.Expected {
				return true, true, submodulePath
			}

			inside, unclean, nestedSubmodulePath := sr.CheckIfFilePathInsideSubmodule(relPath)
			if inside {
				return true, unclean, nestedSubmodulePath
			}

			return true, false, submodulePath
		}
	}

	return false, false, ""
}

func (r *Result) IsFileModified(relPath string, options FilterOptions) bool {
	for _, filePath := range r.filteredFilePathList(options) {
		if path.Join(r.repositoryFullFilepath, filePath) == filepath.ToSlash(relPath) {
			return true
		}
	}

	for _, sr := range r.submoduleResults {
		if util.IsSubpathOfBasePath(filepath.ToSlash(sr.repositoryFullFilepath), filepath.ToSlash(relPath)) {
			if sr.Current != sr.Expected {
				return true
			}

			return sr.IsFileModified(relPath, options)
		}
	}

	return false
}

func isFileStatusAccepted(fileStatus *git.FileStatus, options FilterOptions) bool {
	if (options.StagingOnly && !isFileStatusCodeExpected(fileStatus.Staging)) || (options.WorktreeOnly && !isFileStatusCodeExpected(fileStatus.Worktree)) {
		return false
	}

	return true
}

func isFileStatusCodeExpected(code git.StatusCode) bool {
	return code != git.Unmodified
}
