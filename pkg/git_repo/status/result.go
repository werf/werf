package status

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

type Result struct {
	repositoryFullFilepath string // path relative to main repository
	fileStatusList         git.Status
	submoduleResults       []*SubmoduleResult
}

func NewResult(repositoryFullFilepath string, fileStatusList git.Status, submoduleResults []*SubmoduleResult) *Result {
	return &Result{
		repositoryFullFilepath: repositoryFullFilepath,
		fileStatusList:         fileStatusList,
		submoduleResults:       submoduleResults,
	}
}

type SubmoduleResult struct {
	*Result
	submoduleName   string
	submodulePath   string
	submoduleStatus *git.SubmoduleStatus
}

func NewSubmoduleResult(submoduleName, submodulePath string, submoduleStatus *git.SubmoduleStatus, result *Result) *SubmoduleResult {
	return &SubmoduleResult{
		submoduleName:   submoduleName,
		submodulePath:   submodulePath,
		submoduleStatus: submoduleStatus,
		Result:          result,
	}
}

type FilterOptions struct {
	StagingOnly      bool
	WorktreeOnly     bool
	IgnoreSubmodules bool
}

func (r *Result) Status(ctx context.Context, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := NewResult(r.repositoryFullFilepath, git.Status{}, []*SubmoduleResult{})

	for fileStatusPath, fileStatus := range r.fileStatusList {
		fileStatusFilepath := filepath.FromSlash(fileStatusPath)
		fileStatusFullFilepath := filepath.Join(r.repositoryFullFilepath, fileStatusFilepath)

		if pathMatcher.IsPathMatched(fileStatusFullFilepath) {
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
		if pathMatcher.IsDirOrSubmodulePathMatched(submoduleResult.repositoryFullFilepath) {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogF("Submodule was checking: %s\n", submoduleResult.repositoryFullFilepath)
			}

			newResult, err := submoduleResult.Status(ctx, pathMatcher)
			if err != nil {
				return nil, err
			}

			newSubmoduleResult := NewSubmoduleResult(submoduleResult.submoduleName, submoduleResult.submodulePath, submoduleResult.submoduleStatus, newResult)

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

	if !options.IgnoreSubmodules {
		for _, submoduleResult := range r.submoduleResults {
			result = append(result, submoduleResult.FilePathList(options)...)
		}
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

type UncleanSubmoduleError struct {
	SubmodulePath           string
	ExpectedCommit          string
	CurrentCommit           string
	HeadCommitCurrentCommit string
	error
}

type SubmoduleHasUncommittedChangesError struct {
	SubmodulePath string
	FilePathList  []string
	error
}

func (r *Result) ValidateSubmodules(repository *git.Repository, headCommit string) error {
	if len(r.submoduleResults) == 0 {
		return nil
	}

	c, err := repository.CommitObject(plumbing.NewHash(headCommit))
	if err != nil {
		return err
	}

	cTree, err := c.Tree()
	if err != nil {
		return err
	}

	for _, sr := range r.submoduleResults {
		dotGitExist, err := util.FileExists(filepath.Join(sr.repositoryFullFilepath, ".git"))
		if err != nil {
			return err
		}

		/* The submodule is not checked out, so it is not modified */
		if !dotGitExist {
			continue
		}

		e, err := cTree.FindEntry(sr.submodulePath)
		if err != nil {
			/* The submodule exists locally but it is not committed yet */
			if err == object.ErrEntryNotFound {
				return UncleanSubmoduleError{
					SubmodulePath:           sr.repositoryFullFilepath,
					HeadCommitCurrentCommit: plumbing.ZeroHash.String(),
					ExpectedCommit:          sr.submoduleStatus.Expected.String(),
					CurrentCommit:           sr.submoduleStatus.Current.String(),
					error:                   fmt.Errorf("submodule is not clean"),
				}
			}

			return err
		}

		headCommitSubmoduleCommit := e.Hash

		/* The submodule is switched to another commit and not committed yet */
		if headCommitSubmoduleCommit != sr.submoduleStatus.Expected || sr.submoduleStatus.Expected != sr.submoduleStatus.Current {
			return UncleanSubmoduleError{
				SubmodulePath:           sr.repositoryFullFilepath,
				HeadCommitCurrentCommit: headCommitSubmoduleCommit.String(),
				ExpectedCommit:          sr.submoduleStatus.Expected.String(),
				CurrentCommit:           sr.submoduleStatus.Current.String(),
				error:                   fmt.Errorf("submodule is not clean"),
			}
		}

		/* The submodule expected commit (from stage) differs from the current commit */
		if sr.submoduleStatus.Expected != sr.submoduleStatus.Current {
			/* skip invalid submodule state */
			if sr.submoduleStatus.Current == plumbing.ZeroHash {
				continue
			}

			return UncleanSubmoduleError{
				SubmodulePath:           sr.repositoryFullFilepath,
				HeadCommitCurrentCommit: headCommitSubmoduleCommit.String(),
				ExpectedCommit:          sr.submoduleStatus.Expected.String(),
				CurrentCommit:           sr.submoduleStatus.Current.String(),
				error:                   fmt.Errorf("submodule is not clean"),
			}
		}

		/* The submodule has untracked/modified files */
		if len(sr.fileStatusList) != 0 {
			return SubmoduleHasUncommittedChangesError{
				SubmodulePath: sr.repositoryFullFilepath,
				FilePathList:  sr.filteredFilePathList(FilterOptions{IgnoreSubmodules: true}),
				error:         fmt.Errorf("submodule has uncommitted changes"),
			}
		}

		w, err := repository.Worktree()
		if err != nil {
			return err
		}

		s, err := w.Submodule(sr.submoduleName)
		if err != nil {
			return err
		}

		srRepository, err := s.Repository()
		if err != nil {
			return err
		}

		if err := sr.ValidateSubmodules(srRepository, sr.submoduleStatus.Current.String()); err != nil {
			return err
		}
	}

	return nil
}

func (r *Result) IsFileModified(relPath string, options FilterOptions) bool {
	for _, filePath := range r.filteredFilePathList(options) {
		if path.Join(r.repositoryFullFilepath, filePath) == filepath.ToSlash(relPath) || util.IsSubpathOfBasePath(relPath, filePath) {
			return true
		}
	}

	if !options.IgnoreSubmodules {
		for _, sr := range r.submoduleResults {
			if util.IsSubpathOfBasePath(sr.repositoryFullFilepath, relPath) {
				if sr.submoduleStatus.Current != sr.submoduleStatus.Expected {
					return true
				}

				return sr.IsFileModified(relPath, options)
			}
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
