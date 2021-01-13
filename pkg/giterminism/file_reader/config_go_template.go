package file_reader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/werf/pkg/giterminism"
)

func (r FileReader) ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	err := r.configurationFilesGlob(
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
		func(relPath string) error {
			return giterminism.NewUncommittedConfigurationError(fmt.Sprintf("{{ .Files.Glob '%s' }}: the file '%s' must be committed", pattern, filepath.FromSlash(relPath)))
		},
	)

	return result, err
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
			return giterminism.NewError(fmt.Sprintf("the file '%s' must be committed", filepath.FromSlash(relPath)))
		}
	} else {
		if shouldReadFromFS {
			return fmt.Errorf("the file '%s' not found in the project directory", filepath.FromSlash(relPath))
		} else {
			return giterminism.NewError(fmt.Sprintf("the file '%s' not found in the project git repository", filepath.ToSlash(relPath)))
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
	return r.readCommitFile(ctx, relPath, func(ctx context.Context, s string) error {
		return giterminism.NewUncommittedConfigurationError(fmt.Sprintf("the file '%s' must be committed", filepath.FromSlash(relPath)))
	})
}
