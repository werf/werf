package file_reader

import (
	"context"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo"
)

// WalkConfigurationFilesWithGlob reads the configuration files taking into account the giterminism config.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) WalkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, isFileAcceptedCheckFunc func(relPath string) (bool, error), handleFileFunc func(notResolvedPath string, data []byte, err error) error) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ConfigurationFilesGlob %q %q", dir, glob).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.walkConfigurationFilesWithGlob(ctx, dir, glob, isFileAcceptedCheckFunc, handleFileFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) walkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, isFileAcceptedCheckFunc func(relPath string) (bool, error), handleFileFunc func(notResolvedPath string, data []byte, err error) error) (err error) {
	processedFiles := map[string]bool{}

	isFileProcessedFunc := func(relPath string) bool {
		return processedFiles[filepath.ToSlash(relPath)]
	}

	readFileBeforeHookFunc := func(relPath string) {
		processedFiles[filepath.ToSlash(relPath)] = true
	}

	readFileFunc := func(relPath string) ([]byte, error) {
		readFileBeforeHookFunc(relPath)
		return r.ReadFile(ctx, relPath)
	}

	readCommitFileWrapperFunc := func(relPath string) ([]byte, error) {
		readFileBeforeHookFunc(relPath)
		return r.ReadAndValidateCommitFile(ctx, relPath)
	}

	fileRelPathListFromFS, err := r.ListFilesWithGlob(ctx, dir, glob)
	if err != nil {
		return err
	}

	if r.sharedOptions.LooseGiterminism() {
		for _, relPath := range fileRelPathListFromFS {
			data, err := readFileFunc(relPath)
			if err := handleFileFunc(relPath, data, err); err != nil {
				return err
			}
		}

		return nil
	}

	fileRelPathListFromCommit, err := r.ListCommitFilesWithGlob(ctx, dir, glob)
	if err != nil {
		return err
	}

	var relPathListWithUncommittedFilesChanges []string
	for _, relPath := range fileRelPathListFromCommit {
		if accepted, err := r.IsFilePathAccepted(ctx, relPath, isFileAcceptedCheckFunc); err != nil {
			return err
		} else if accepted {
			continue
		}

		data, err := readCommitFileWrapperFunc(relPath)
		if err := handleFileFunc(relPath, data, err); err != nil {
			if isUncommittedFilesChangesError(err) {
				relPathListWithUncommittedFilesChanges = append(relPathListWithUncommittedFilesChanges, relPath)
				continue
			}

			return err
		}
	}

	if len(relPathListWithUncommittedFilesChanges) != 0 {
		return NewUncommittedFilesChangesError(relPathListWithUncommittedFilesChanges...)
	}

	var relPathListWithUncommittedFiles []string
	for _, relPath := range fileRelPathListFromFS {
		accepted, err := r.IsFilePathAccepted(ctx, relPath, isFileAcceptedCheckFunc)
		if err != nil {
			return err
		}

		if !accepted {
			if !isFileProcessedFunc(relPath) {
				relPathListWithUncommittedFiles = append(relPathListWithUncommittedFiles, relPath)
			}

			continue
		}

		data, err := readFileFunc(relPath)
		if err := handleFileFunc(relPath, data, err); err != nil {
			return err
		}
	}

	if len(relPathListWithUncommittedFiles) != 0 {
		return NewUncommittedFilesError(relPathListWithUncommittedFiles...)
	}

	return nil
}

func (r FileReader) ReadAndValidateConfigurationFile(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadAndValidateConfigurationFile %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readAndValidateConfigurationFile(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readAndValidateConfigurationFile(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) ([]byte, error) {
	existAndAccepted, err := r.IsRegularFileExistAndAccepted(ctx, relPath, isFileAcceptedCheckFunc)
	if err != nil {
		return nil, err
	}

	if existAndAccepted {
		return r.ReadFile(ctx, relPath)
	}

	return r.ReadAndValidateCommitFile(ctx, relPath)
}

// CheckConfigurationFileExistence returns an error if there are not the file in the project repository commit and an accepted file by the giteminism config in the project directory.
func (r FileReader) CheckConfigurationFileExistence(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("CheckConfigurationFileExistence %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.checkConfigurationFileExistence(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) checkConfigurationFileExistence(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) error {
	existInFS, err := r.IsRegularFileExist(ctx, relPath)
	if err != nil {
		return err
	}

	if existInFS {
		if r.sharedOptions.LooseGiterminism() {
			return nil
		}

		accepted, err := r.IsFilePathAccepted(ctx, relPath, isFileAcceptedCheckFunc)
		if err != nil {
			return err
		}

		if accepted {
			return nil
		}
	}

	if r.sharedOptions.LooseGiterminism() {
		return NewFilesNotFoundInProjectDirectoryError(relPath)
	}

	exist, err := r.IsCommitFileExist(ctx, relPath)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	if existInFS {
		err := r.ValidateCommitFilePath(ctx, relPath)
		if git_repo.IsTreeEntryNotFoundInRepoErr(err) {
			return NewUncommittedFilesError(relPath)
		} else {
			return err
		}
	} else {
		return NewFilesNotFoundInProjectGitRepositoryError(relPath)
	}
}

// IsConfigurationFileExistAnywhere checks the configuration file existence in the project directory and the project repository.
func (r FileReader) IsConfigurationFileExistAnywhere(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsConfigurationFileExistAnywhere %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.isConfigurationFileExistAnywhere(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isConfigurationFileExistAnywhere(ctx context.Context, relPath string) (bool, error) {
	exist, err := r.IsRegularFileExist(ctx, relPath)
	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
	}

	if r.sharedOptions.LooseGiterminism() {
		return false, nil
	} else {
		return r.IsCommitFileExist(ctx, relPath)
	}
}

// IsConfigurationFileExist checks the configuration file existence taking into account the giterminism config.
// The method applies isFileAcceptedCheckFunc for each resolved path.
func (r FileReader) IsConfigurationFileExist(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsConfigurationFileExist %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.isConfigurationFileExist(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isConfigurationFileExist(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (bool, error) {
	exist, err := r.IsRegularFileExistAndAccepted(ctx, relPath, isFileAcceptedCheckFunc)
	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
	}

	if r.sharedOptions.LooseGiterminism() {
		return false, nil
	} else {
		return r.IsCommitFileExist(ctx, relPath)
	}
}
