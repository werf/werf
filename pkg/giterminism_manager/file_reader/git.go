package file_reader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo/status"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func (r FileReader) toWorkTreeRelativePath(relPath string) string {
	return filepath.Join(r.sharedOptions.RelativeToGitProjectDir(), relPath)
}

func (r FileReader) isSubpathOfWorkTreeDir(absPath string) bool {
	return util.IsSubpathOfBasePath(r.sharedOptions.LocalGitRepo().WorkTreeDir, absPath)
}

func (r FileReader) ValidateRelatedSubmodules(ctx context.Context, relPath string) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ValidateRelatedSubmodules %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.validateRelatedSubmodules(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) validateRelatedSubmodules(ctx context.Context, relPath string) error {
	return r.sharedOptions.LocalGitRepo().ValidateSubmodules(ctx, path_matcher.NewSimplePathMatcher(r.toWorkTreeRelativePath(relPath), nil))
}

func (r FileReader) IsCommitFileModifiedLocally(ctx context.Context, relPath string) (modified bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitFileModifiedLocally %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			modified, err = r.isCommitFileModifiedLocally(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("modified: %v\nerr: %q\n", modified, err)
			}
		})

	return
}

func (r FileReader) isCommitFileModifiedLocally(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsFileModifiedLocally(ctx, r.toWorkTreeRelativePath(relPath), status.FilterOptions{
		WorktreeOnly:     r.sharedOptions.Dev(),
		IgnoreSubmodules: true,
	})
}

func (r FileReader) IsCommitFileExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitFileExist(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(relPath))
}

func (r FileReader) IsCommitTreeEntryExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitTreeEntryExist(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(relPath))
}

func (r FileReader) IsCommitTreeEntryDirectory(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitTreeEntryDirectory(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(relPath))
}

func (r FileReader) ReadCommitTreeEntryContent(ctx context.Context, relPath string) ([]byte, error) {
	return r.sharedOptions.LocalGitRepo().ReadCommitTreeEntryContent(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(relPath))
}

func (r FileReader) ResolveAndCheckCommitFilePath(ctx context.Context, relPath string, checkSymlinkTargetFunc func(resolvedRelPath string) error) (string, error) {
	return r.sharedOptions.LocalGitRepo().ResolveAndCheckCommitFilePath(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(relPath), checkSymlinkTargetFunc)
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
	list, err := r.sharedOptions.LocalGitRepo().ListCommitFilesWithGlob(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(dir), pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, path := range list {
		relativeToGitProjectDirPath := util.GetRelativeToBaseFilepath(r.sharedOptions.RelativeToGitProjectDir(), path)
		relativeToDirPath := util.GetRelativeToBaseFilepath(dir, relativeToGitProjectDirPath)
		result = append(result, relativeToDirPath)
	}

	return result, nil
}

func (r FileReader) ReadCommitFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.sharedOptions.LocalGitRepo().ReadCommitFile(ctx, r.sharedOptions.HeadCommit(), r.toWorkTreeRelativePath(relPath))
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
			resolvedRelPathRelativeToProjectDir := util.GetRelativeToBaseFilepath(r.sharedOptions.RelativeToGitProjectDir(), resolvedRelPath)

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

// IsFileModifiedLocally checks if the file modified locally, not ignored by .gitignore, or inside an unclean submodule repository.
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
	err := r.checkFileModifiedLocally(ctx, relPath)
	if err != nil {
		switch err.(type) {
		case UncommittedFilesError:
			return true, nil
		default:
			return false, err
		}
	}

	return false, nil
}

func (r FileReader) checkFileModifiedLocally(ctx context.Context, relPath string) error {
	if err := r.ValidateRelatedSubmodules(ctx, relPath); err != nil {
		return r.HandleValidateSubmodulesErr(err)
	}

	isFileModified, err := r.IsCommitFileModifiedLocally(ctx, relPath)
	if err != nil {
		return err
	}

	if !isFileModified {
		return nil
	}

	return r.extraCheckFileModifiedLocally(ctx, relPath)
}

func (r FileReader) HandleValidateSubmodulesErr(err error) error {
	switch statusErr := err.(type) {
	case status.UncleanSubmoduleError:
		return r.NewUncleanSubmoduleError(statusErr.SubmodulePath, statusErr.HeadCommitCurrentCommit, statusErr.CurrentCommit, statusErr.ExpectedCommit)
	case status.SubmoduleHasUncommittedChangesError:
		return r.NewUncommittedSubmoduleChangesError(statusErr.SubmodulePath, statusErr.FilePathList)
	default:
		return err
	}
}

// https://github.com/go-git/go-git/issues/227
// https://github.com/go-git/go-git/issues/253
func (r FileReader) ExtraCheckFilesModifiedLocally(ctx context.Context, relPathList ...string) error {
	var uncommittedFilePathList []string

	for _, relPath := range relPathList {
		err := r.extraCheckFileModifiedLocally(ctx, relPath)
		if err != nil {
			switch err.(type) {
			case UncommittedFilesError:
				uncommittedFilePathList = append(uncommittedFilePathList, relPath)
				continue
			}

			return err
		}
	}

	if len(uncommittedFilePathList) != 0 {
		return r.NewUncommittedFilesError(uncommittedFilePathList...)
	}

	return nil
}

func (r FileReader) extraCheckFileModifiedLocally(ctx context.Context, relPath string) error {
	isTreeEntryExist, err := r.IsCommitTreeEntryExist(ctx, relPath)
	if err != nil {
		return err
	}

	if !isTreeEntryExist {
		return r.NewUncommittedFilesError(relPath)
	}

	isCommitTreeEntryDirectory, err := r.IsCommitTreeEntryDirectory(ctx, relPath)
	if err != nil {
		return err
	}

	if isCommitTreeEntryDirectory {
		return r.NewUncommittedFilesError(relPath)
	}

	commitFileData, err := r.ReadCommitTreeEntryContent(ctx, relPath)
	if err != nil {
		return err
	}

	absPath := r.toProjectDirAbsolutePath(relPath)
	isFileExist, err := util.FileExists(absPath)
	if err != nil {
		return err
	}

	if !isFileExist {
		return r.NewUncommittedFilesError(relPath)
	}

	lstat, err := os.Lstat(absPath)
	if err != nil {
		return err
	}

	if lstat.IsDir() {
		return r.NewUncommittedFilesError(relPath)
	}

	var fsFileData []byte
	if lstat.Mode()&os.ModeSymlink == os.ModeSymlink {
		link, err := os.Readlink(absPath)
		if err != nil {
			return err
		}

		fsFileData = []byte(link)
	} else {
		data, err := r.ReadFile(ctx, relPath)
		if err != nil {
			return err
		}

		fsFileData = data
	}

	isDataIdentical := bytes.Equal(commitFileData, fsFileData)
	localDataWithForcedUnixLineBreak := bytes.ReplaceAll(fsFileData, []byte("\r\n"), []byte("\n"))
	if !isDataIdentical {
		isDataIdentical = bytes.Equal(commitFileData, localDataWithForcedUnixLineBreak)
	}

	if !isDataIdentical {
		return r.NewUncommittedFilesError(relPath)
	}

	return nil
}
