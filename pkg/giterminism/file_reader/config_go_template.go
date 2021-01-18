package file_reader

import (
	"context"
	"fmt"
	"path/filepath"
)

func (r FileReader) ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if err := r.configurationFilesGlob(
		ctx,
		pattern,
		r.manager.Config().IsUncommittedConfigGoTemplateRenderingFileAccepted,
		r.readCommitConfigGoTemplateFile,
		func(relPath string, data []byte, err error) error {
			if err != nil {
				return err
			}

			result[filepath.ToSlash(relPath)] = string(data)

			return nil
		},
		func(relPaths ...string) error {
			return NewUncommittedFilesError("file", relPaths...)
		},
		func(relPaths ...string) error {
			return NewUncommittedFilesChangesError("file", relPaths...)
		},
	); err != nil {
		return nil, fmt.Errorf("{{ .Files.Glob '%s' }}: %s", pattern, err)
	}

	return result, nil
}

func (r FileReader) ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error) {
	if err := r.checkConfigGoTemplateFileExistence(ctx, relPath); err != nil {
		return nil, fmt.Errorf("{{ .Files.Get '%s' }}: %s", relPath, err)
	}

	data, err := r.readConfigGoTemplateFile(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("{{ .Files.Get '%s' }}: %s", relPath, err)
	}

	return data, nil
}

func (r FileReader) checkConfigGoTemplateFileExistence(ctx context.Context, relPath string) error {
	accepted, err := r.manager.Config().IsUncommittedConfigGoTemplateRenderingFileAccepted(relPath)
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
			return NewUncommittedFilesError("file", relPath)
		}
	} else {
		if shouldReadFromFS {
			return NewFilesNotFoundInTheProjectDirectoryError("file", relPath)
		} else {
			return NewFilesNotFoundInTheProjectGitRepositoryError("file", relPath)
		}
	}
}

func (r FileReader) readConfigGoTemplateFile(ctx context.Context, relPath string) ([]byte, error) {
	accepted, err := r.manager.Config().IsUncommittedConfigGoTemplateRenderingFileAccepted(relPath)
	if err != nil {
		return nil, err
	}

	if r.manager.LooseGiterminism() || accepted {
		return r.readFile(relPath)
	}

	return r.readCommitConfigGoTemplateFile(ctx, relPath)
}

func (r FileReader) readCommitConfigGoTemplateFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.readCommitFile(ctx, relPath, func(ctx context.Context, relPath string) error {
		return NewUncommittedFilesChangesError("file", relPath)
	})
}
