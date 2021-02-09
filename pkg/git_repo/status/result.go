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
	repository             *git.Repository
	repositoryAbsFilepath  string // absolute path
	repositoryFullFilepath string // path relative to main repository
	fileStatusList         git.Status
	submoduleResults       []*SubmoduleResult
}

type SubmoduleResult struct {
	*Result
	SubmodulePath string
	*git.SubmoduleStatus
}

type FilterOptions struct {
	StagingOnly      bool
	WorktreeOnly     bool
	IgnoreSubmodules bool
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
				SubmodulePath:   submoduleResult.SubmodulePath,
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

func (r *Result) ValidateSubmodules(headCommit string) error {
	if len(r.submoduleResults) == 0 {
		return nil
	}

	c, err := r.repository.CommitObject(plumbing.NewHash(headCommit))
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

		e, err := cTree.FindEntry(sr.SubmodulePath)
		if err != nil {
			/* The submodule exists locally but it is not committed yet */
			if err == object.ErrEntryNotFound {
				return UncleanSubmoduleError{
					SubmodulePath:           sr.repositoryFullFilepath,
					HeadCommitCurrentCommit: plumbing.ZeroHash.String(),
					ExpectedCommit:          sr.Expected.String(),
					CurrentCommit:           sr.Current.String(),
					error:                   fmt.Errorf("submodule is not clean"),
				}
			}

			return err
		}

		headCommitSubmoduleCommit := e.Hash

		/* The submodule is switched to another commit and not committed yet */
		if headCommitSubmoduleCommit != sr.Expected {
			return UncleanSubmoduleError{
				SubmodulePath:           sr.repositoryFullFilepath,
				HeadCommitCurrentCommit: headCommitSubmoduleCommit.String(),
				ExpectedCommit:          sr.Expected.String(),
				CurrentCommit:           sr.Current.String(),
				error:                   fmt.Errorf("submodule is not clean"),
			}
		}

		/* The submodule expected commit (from stage) differs from the current commit */
		if sr.Expected != sr.Current {
			/* skip invalid submodule state */
			if sr.Current == plumbing.ZeroHash {
				continue
			}

			return UncleanSubmoduleError{
				SubmodulePath:           sr.repositoryFullFilepath,
				HeadCommitCurrentCommit: headCommitSubmoduleCommit.String(),
				ExpectedCommit:          sr.Expected.String(),
				CurrentCommit:           sr.Current.String(),
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

		if err := sr.ValidateSubmodules(sr.Current.String()); err != nil {
			return err
		}
	}

	return nil
}

func (r *Result) IsFileModified(relPath string, options FilterOptions) bool {
	for _, filePath := range r.filteredFilePathList(options) {
		if path.Join(r.repositoryFullFilepath, filePath) == filepath.ToSlash(relPath) {
			return true
		}
	}

	if !options.IgnoreSubmodules {
		for _, sr := range r.submoduleResults {
			if util.IsSubpathOfBasePath(filepath.ToSlash(sr.repositoryFullFilepath), filepath.ToSlash(relPath)) {
				if sr.Current != sr.Expected {
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
