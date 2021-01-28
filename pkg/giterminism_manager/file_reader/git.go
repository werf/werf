package file_reader

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
)

func (r FileReader) relativeToGitPath(relPath string) string {
	return filepath.Join(r.sharedOptions.RelativeToGitProjectDir(), relPath)
}

func (r FileReader) IsCommitFileExist(ctx context.Context, relPath string) (bool, error) {
	return r.sharedOptions.LocalGitRepo().IsCommitFileExist(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath))
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

func (r FileReader) ResolveAndCheckCommitFilePath(ctx context.Context, relPath string, checkFunc func(resolvedRelPath string) error) (string, error) {
	return r.sharedOptions.LocalGitRepo().ResolveAndCheckCommitFilePath(ctx, r.sharedOptions.HeadCommit(), r.relativeToGitPath(relPath), checkFunc)
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
		_, err := r.readCommitFile(ctx, relPath)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r FileReader) readCommitFile(ctx context.Context, relPath string) ([]byte, error) {
	logboek.Context(ctx).Debug().LogF("-- giterminism_manager.FileReader.readCommitFile relPath=%q\n", relPath)

	repoData, err := r.ReadCommitFile(ctx, relPath)
	if err != nil {
		return nil, err
	}

	isDataIdentical, err := r.compareFileData(ctx, relPath, repoData)
	if err != nil {
		return nil, fmt.Errorf("unable to compare commit file %q with the local project file: %s", relPath, err)
	}

	if !isDataIdentical {
		if err := NewUncommittedFilesChangesError(relPath); err != nil {
			return nil, err
		}
	}

	return repoData, err
}

func (r FileReader) compareFileData(ctx context.Context, relPath string, data []byte) (bool, error) {
	var fileData []byte
	if exist, err := r.IsRegularFileExist(ctx, relPath); err != nil {
		return false, err
	} else if exist {
		fileData, err = r.ReadFile(ctx, relPath)
		if err != nil {
			return false, err
		}
	}

	isDataIdentical := bytes.Equal(fileData, data)
	fileDataWithForcedUnixLineBreak := bytes.ReplaceAll(fileData, []byte("\r\n"), []byte("\n"))
	if !isDataIdentical {
		isDataIdentical = bytes.Equal(fileDataWithForcedUnixLineBreak, data)
	}

	return isDataIdentical, nil
}
