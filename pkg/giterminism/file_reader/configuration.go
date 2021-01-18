package file_reader

import (
	"context"
	"path/filepath"
)

func (r FileReader) configurationFilesGlob(ctx context.Context, configType configType, pattern string, isFileAcceptedFunc func(relPath string) (bool, error), readCommitFileFunc func(ctx context.Context, relPath string) ([]byte, error), handleFileFunc func(relPath string, data []byte, err error) error) error {
	processedFiles := map[string]bool{}

	isFileProcessedFunc := func(relPath string) bool {
		return processedFiles[filepath.ToSlash(relPath)]
	}

	readFileBeforeHookFunc := func(relPath string) {
		processedFiles[filepath.ToSlash(relPath)] = true
	}

	readFileFunc := func(relPath string) ([]byte, error) {
		readFileBeforeHookFunc(relPath)
		return r.readFile(relPath)
	}

	readCommitFileWrapperFunc := func(relPath string) ([]byte, error) {
		readFileBeforeHookFunc(relPath)
		return readCommitFileFunc(ctx, relPath)
	}

	fileRelPathListFromFS, err := r.filesGlob(pattern)
	if err != nil {
		return err
	}

	if r.manager.LooseGiterminism() {
		for _, relPath := range fileRelPathListFromFS {
			data, err := readFileFunc(relPath)
			if err := handleFileFunc(relPath, data, err); err != nil {
				return err
			}
		}

		return nil
	}

	fileRelPathListFromCommit, err := r.commitFilesGlob(ctx, pattern)
	if err != nil {
		return err
	}

	var relPathListWithUncommittedFilesChanges []string
	for _, relPath := range fileRelPathListFromCommit {
		if accepted, err := isFileAcceptedFunc(relPath); err != nil {
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
		return NewUncommittedFilesChangesError(configType, relPathListWithUncommittedFilesChanges...)
	}

	var relPathListWithUncommittedFiles []string
	for _, relPath := range fileRelPathListFromFS {
		accepted, err := isFileAcceptedFunc(relPath)
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
		return NewUncommittedFilesError(configType, relPathListWithUncommittedFiles...)
	}

	return nil
}

func (r FileReader) readConfigurationFile(ctx context.Context, configType configType, relPath string, isFileAcceptedFunc func(relPath string) (bool, error)) ([]byte, error) {
	accepted, err := isFileAcceptedFunc(relPath)
	if err != nil {
		return nil, err
	}

	if r.manager.LooseGiterminism() || accepted {
		return r.readFile(relPath)
	}

	return r.readCommitFile(ctx, relPath, func(ctx context.Context, relPath string) error {
		return NewUncommittedFilesChangesError(configType, relPath)
	})
}

func (r FileReader) checkConfigurationDirectoryExistence(ctx context.Context, configType configType, relPath string, isFileAcceptedFunc func(relPath string) (bool, error)) error {
	accepted, err := isFileAcceptedFunc(relPath)
	if err != nil {
		return err
	}

	shouldReadFromFS := r.manager.LooseGiterminism() || accepted
	if !shouldReadFromFS {
		if exist, err := r.isCommitDirectoryExist(ctx, relPath); err != nil {
			return err
		} else if exist {
			return nil
		}
	}

	exist, err := r.isDirectoryExist(relPath)
	if err != nil {
		return err
	}

	if exist {
		if shouldReadFromFS {
			return nil
		} else {
			return NewUncommittedFilesError(configType, relPath)
		}
	} else {
		if shouldReadFromFS {
			return NewFilesNotFoundInTheProjectDirectoryError(configType, relPath)
		} else {
			return NewFilesNotFoundInTheProjectGitRepositoryError(configType, relPath)
		}
	}
}

func (r FileReader) checkConfigurationFileExistence(ctx context.Context, configType configType, relPath string, isFileAcceptedFunc func(relPath string) (bool, error)) error {
	accepted, err := isFileAcceptedFunc(relPath)
	if err != nil {
		return err
	}

	shouldReadFromFS := r.manager.LooseGiterminism() || accepted
	if !shouldReadFromFS {
		if exist, err := r.isCommitFileExist(ctx, relPath); err != nil {
			return err
		} else if exist {
			return nil
		}
	}

	exist, err := r.isFileExist(relPath)
	if err != nil {
		return err
	}

	if exist {
		if shouldReadFromFS {
			return nil
		} else {
			return NewUncommittedFilesError(configType, relPath)
		}
	} else {
		if shouldReadFromFS {
			return NewFilesNotFoundInTheProjectDirectoryError(configType, relPath)
		} else {
			return NewFilesNotFoundInTheProjectGitRepositoryError(configType, relPath)
		}
	}
}

func (r FileReader) isConfigurationFileExistAnywhere(ctx context.Context, relPath string) (bool, error) {
	if exist, err := r.isCommitFileExist(ctx, relPath); err != nil {
		return false, err
	} else if !exist {
		return r.isFileExist(relPath)
	} else {
		return true, nil
	}
}

func (r FileReader) isConfigurationDirectoryExist(ctx context.Context, relPath string, isFileAcceptedFunc func(relPath string) (bool, error)) (bool, error) {
	accepted, err := isFileAcceptedFunc(relPath)
	if err != nil {
		return false, err
	}

	if r.manager.LooseGiterminism() || accepted {
		return r.isDirectoryExist(relPath)
	}

	return r.isCommitDirectoryExist(ctx, relPath)
}

func (r FileReader) isConfigurationFileExist(ctx context.Context, relPath string, isFileAcceptedFunc func(relPath string) (bool, error)) (bool, error) {
	accepted, err := isFileAcceptedFunc(relPath)
	if err != nil {
		return false, err
	}

	if r.manager.LooseGiterminism() || accepted {
		return r.isFileExist(relPath)
	}

	return r.isCommitFileExist(ctx, relPath)
}
