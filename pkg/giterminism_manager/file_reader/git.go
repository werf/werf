package file_reader

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func (r FileReader) projectDirRelativePathToWorkTreeRelativePath(relPath string) string {
	return filepath.Join(r.sharedOptions.RelativeToGitProjectDir(), relPath)
}

func (r FileReader) gitRelativePathToProjectDirRelativePath(relToGitPath string) string {
	return util.GetRelativeToBaseFilepath(r.sharedOptions.RelativeToGitProjectDir(), relToGitPath)
}

func (r FileReader) isSubpathOfWorkTreeDir(absPath string) bool {
	return util.IsSubpathOfBasePath(r.sharedOptions.LocalGitRepo().WorkTreeDir, absPath)
}

func (r FileReader) ValidateSubmodules(ctx context.Context, pathMatcher path_matcher.PathMatcher) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ValidateSubmodules %q", pathMatcher.String()).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.validateSubmodules(ctx, pathMatcher)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) validateSubmodules(ctx context.Context, pathMatcher path_matcher.PathMatcher) error {
	return r.sharedOptions.LocalGitRepo().ValidateSubmodules(ctx, pathMatcher, git_repo.ValidateSubmodulesOptions{OnlyWorktreeChanges: r.sharedOptions.Dev()})
}

func (r FileReader) StatusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher) (list []string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("StatusPathList %q", pathMatcher.String()).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			list, err = r.statusPathList(ctx, pathMatcher)

			if debug() {
				logboek.Context(ctx).Debug().LogF("list: %v\nerr: %q\n", list, err)
			}
		})

	return
}

func (r FileReader) statusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher) ([]string, error) {
	list, err := r.sharedOptions.LocalGitRepo().StatusPathList(ctx, pathMatcher, git_repo.StatusPathListOptions{OnlyWorktreeChanges: r.sharedOptions.Dev()})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, relPath := range list {
		result = append(result, r.gitRelativePathToProjectDirRelativePath(relPath))
	}

	return result, nil
}

func (r FileReader) IsCommitFileExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitFileExist(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(relPath))
}

func (r FileReader) IsCommitTreeEntryExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitTreeEntryExist(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(relPath))
}

func (r FileReader) IsCommitTreeEntryDirectory(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitTreeEntryDirectory(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(relPath))
}

func (r FileReader) ReadCommitTreeEntryContent(ctx context.Context, relPath string) ([]byte, error) {
	return r.sharedOptions.LocalGitRepo().ReadCommitTreeEntryContent(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(relPath))
}

func (r FileReader) ResolveAndCheckCommitFilePath(ctx context.Context, relPath string, checkSymlinkTargetFunc func(resolvedRelPath string) error) (string, error) {
	return r.sharedOptions.LocalGitRepo().ResolveAndCheckCommitFilePath(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(relPath), checkSymlinkTargetFunc)
}

func (r FileReader) ListCommitFilesWithGlob(ctx context.Context, dir string, pattern string) (files []string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ListCommitFilesWithGlob %q %q", dir, pattern).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			files, err = r.listCommitFilesWithGlob(ctx, dir, pattern)

			if debug() {
				var logFiles string
				if len(files) != 0 {
					logFiles = fmt.Sprintf("\n - %s", strings.Join(files, "\n - "))
				}
				logboek.Context(ctx).Debug().LogF("files: %v\nerr: %q\n", logFiles, err)
			}
		})

	return
}

func (r FileReader) listCommitFilesWithGlob(ctx context.Context, dir string, pattern string) ([]string, error) {
	list, err := r.sharedOptions.LocalGitRepo().ListCommitFilesWithGlob(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(dir), pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, path := range list {
		relativeToGitProjectDirPath := r.gitRelativePathToProjectDirRelativePath(path)
		relativeToDirPath := util.GetRelativeToBaseFilepath(dir, relativeToGitProjectDirPath)
		result = append(result, relativeToDirPath)
	}

	return result, nil
}

func (r FileReader) ReadCommitFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.sharedOptions.LocalGitRepo().ReadCommitFile(ctx, r.sharedOptions.HeadCommit(), r.projectDirRelativePathToWorkTreeRelativePath(relPath))
}

// CheckCommitFileExistenceAndLocalChanges returns nil if the file exists and does not have any uncommitted changes locally (each symlink target).
func (r FileReader) CheckCommitFileExistenceAndLocalChanges(ctx context.Context, relPath string) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("CheckCommitFileExistenceAndLocalChanges %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.checkCommitFileExistenceAndLocalChanges(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) checkCommitFileExistenceAndLocalChanges(ctx context.Context, relPath string) error {
	if err := r.checkFileModifiedLocally(ctx, relPath); err != nil { // check not resolved path
		return err
	}

	commitTreeEntryExist, err := r.IsCommitTreeEntryExist(ctx, relPath)
	if err != nil {
		return err
	}

	if !commitTreeEntryExist {
		commitFileExist, err := r.IsCommitFileExist(ctx, relPath)
		if err != nil {
			return err
		}

		if !commitFileExist {
			return r.NewFileNotFoundInProjectRepositoryError(relPath)
		}
	}

	if err := func() error {
		resolvedPath, err := r.ResolveAndCheckCommitFilePath(ctx, relPath, func(resolvedRelPath string) error { // check each symlink target
			resolvedRelPathRelativeToProjectDir := r.gitRelativePathToProjectDirRelativePath(resolvedRelPath)

			return r.checkFileModifiedLocally(ctx, resolvedRelPathRelativeToProjectDir)
		})
		if err != nil {
			return err
		}

		if resolvedPath != relPath { // check resolved path
			if err := r.checkFileModifiedLocally(ctx, relPath); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return fmt.Errorf("symlink %q check failed: %s", relPath, err)
	}

	return nil
}

// IsFileModifiedLocally checks for the file changes in worktree or index.
func (r FileReader) IsFileModifiedLocally(ctx context.Context, relPath string) (modified bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsFileModifiedLocally %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			modified, err = r.isFileModifiedLocally(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("modified: %v\nerr: %q\n", modified, err)
			}
		})

	return
}

func (r FileReader) isFileModifiedLocally(ctx context.Context, relPath string) (bool, error) {
	return r.isStatusPathListEmpty(ctx, path_matcher.NewSimplePathMatcher(r.projectDirRelativePathToWorkTreeRelativePath(relPath), nil))
}

func (r FileReader) isStatusPathListEmpty(ctx context.Context, pathMatcher path_matcher.PathMatcher) (bool, error) {
	list, err := r.StatusPathList(ctx, pathMatcher)
	if err != nil {
		return false, err
	}

	return len(list) != 0, nil
}

func (r FileReader) checkFileModifiedLocally(ctx context.Context, relPath string) error {
	return r.ValidateStatusPathList(ctx, path_matcher.NewSimplePathMatcher(r.projectDirRelativePathToWorkTreeRelativePath(relPath), nil))
}

// ValidateStatusPathList returns an error if there are any changes in worktree or index.
func (r FileReader) ValidateStatusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ValidateStatusPathList %q", pathMatcher.String()).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.checkFilesModifiedLocally(ctx, pathMatcher)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) checkFilesModifiedLocally(ctx context.Context, pathMatcher path_matcher.PathMatcher) error {
	list, err := r.StatusPathList(ctx, pathMatcher)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return nil
	}

	if err := r.ValidateSubmodules(ctx, pathMatcher); err != nil {
		return r.HandleValidateSubmodulesErr(err)
	}

	return r.NewUncommittedFilesError(list...)
}

func (r FileReader) HandleValidateSubmodulesErr(err error) error {
	switch statusErr := err.(type) {
	case git_repo.SubmoduleAddedAndNotCommittedError:
		return r.NewSubmoduleAddedAndNotCommittedError(statusErr.SubmodulePath)
	case git_repo.SubmoduleDeletedError:
		return r.NewSubmoduleDeletedError(statusErr.SubmodulePath)
	case git_repo.SubmoduleHasUntrackedChangesError:
		return r.NewSubmoduleHasUntrackedChangesError(statusErr.SubmodulePath)
	case git_repo.SubmoduleHasUncommittedChangesError:
		return r.NewSubmoduleHasUncommittedChangesError(statusErr.SubmodulePath)
	case git_repo.SubmoduleCommitChangedError:
		return r.NewSubmoduleCommitChangedError(statusErr.SubmodulePath)
	default:
		return err
	}
}
