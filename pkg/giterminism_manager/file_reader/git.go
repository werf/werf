package file_reader

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/util"
)

func (r FileReader) relativeToGitPath(relPath string) string {
	return filepath.Join(r.sharedOptions.RelativeToGitProjectDir(), relPath)
}

func (r FileReader) IsCommitFileModifiedLocally(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsFileModifiedLocally(ctx, r.relativeToGitPath(relPath), git_repo.IsFileModifiedLocally{
		WorktreeOnly: r.sharedOptions.Dev(),
	})
}

func (r FileReader) IsCommitFileExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitFileExist(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
}

func (r FileReader) IsCommitTreeEntryExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitTreeEntryExist(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
}

func (r FileReader) ReadCommitTreeEntryContent(ctx context.Context, relPath string) ([]byte, error) {
	return r.sharedOptions.LocalGitRepo().ReadCommitTreeEntryContent(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
}

func (r FileReader) ResolveAndCheckCommitFilePath(ctx context.Context, relPath string, checkFunc func(resolvedRelPath string) error) (string, error) {
	return r.sharedOptions.LocalGitRepo().ResolveAndCheckCommitFilePath(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath), checkFunc)
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
	return r.sharedOptions.LocalGitRepo().ListCommitFilesWithGlob(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(dir), pattern)
}

func (r FileReader) ReadCommitFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.sharedOptions.LocalGitRepo().ReadCommitFile(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
}

func (r FileReader) ReadAndValidateCommitFile(ctx context.Context, relPath string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadAndValidateCommitFile %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readAndValidateCommitConfigurationFile(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readAndValidateCommitConfigurationFile(ctx context.Context, relPath string) ([]byte, error) {
	if err := r.ValidateCommitFilePath(ctx, relPath); err != nil {
		return nil, err
	}

	return r.ReadCommitFile(ctx, relPath)
}

func (r FileReader) ValidateCommitFilePath(ctx context.Context, relPath string) error {
	if _, err := r.ResolveAndCheckCommitFilePath(ctx, relPath, func(resolvedRelPath string) error {
		resolvedRelPathRelativeToProjectDir := util.GetRelativeToBaseFilepath(r.sharedOptions.RelativeToGitProjectDir(), resolvedRelPath)

		isFileModified, err := r.IsCommitFileModifiedLocally(ctx, resolvedRelPathRelativeToProjectDir)
		if err != nil {
			return err
		}

		if !isFileModified {
			return nil
		}

		if runtime.GOOS == "windows" {
			return r.ExtraWindowsCheckFileModifiedLocally(ctx, resolvedRelPathRelativeToProjectDir)
		}

		isTreeEntryExist, err := r.IsCommitTreeEntryExist(ctx, resolvedRelPathRelativeToProjectDir)
		if err != nil {
			return err
		}

		if isTreeEntryExist {
			return NewUncommittedFilesChangesError(resolvedRelPathRelativeToProjectDir)
		} else {
			return NewUncommittedFilesError(resolvedRelPathRelativeToProjectDir)
		}
	}); err != nil {
		return err
	}

	return nil
}

// https://github.com/go-git/go-git/issues/227
func (r FileReader) ExtraWindowsCheckFilesModifiedLocally(ctx context.Context, relPathList ...string) error {
	var uncommittedFilePathList []string

	for _, relPath := range relPathList {
		err := r.ExtraWindowsCheckFileModifiedLocally(ctx, relPath)
		if err != nil {
			switch err.(type) {
			case UncommittedFilesError, UncommittedFilesChangesError:
				uncommittedFilePathList = append(uncommittedFilePathList, relPath)
				continue
			}

			return err
		}
	}

	if len(uncommittedFilePathList) != 0 {
		return NewUncommittedFilesChangesError(uncommittedFilePathList...)
	}

	return nil
}

// https://github.com/go-git/go-git/issues/227
func (r FileReader) ExtraWindowsCheckFileModifiedLocally(ctx context.Context, relPath string) error {
	isTreeEntryExist, err := r.IsCommitTreeEntryExist(ctx, relPath)
	if err != nil {
		return err
	}

	var commitFileData []byte
	if isTreeEntryExist {
		data, err := r.ReadCommitTreeEntryContent(ctx, relPath)
		if err != nil {
			return err
		}

		commitFileData = data
	}

	isFileExist, err := r.IsRegularFileExist(ctx, relPath)
	if err != nil {
		return err
	}

	var fsFileData []byte
	if isFileExist {
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

	if isDataIdentical {
		return nil
	}

	if isTreeEntryExist {
		return NewUncommittedFilesChangesError(relPath)
	} else {
		return NewUncommittedFilesError(relPath)
	}
}
