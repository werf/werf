package file_reader

import (
	"context"
	"fmt"
	"path/filepath"
)

func (r FileReader) ConfigGoTemplateFilesGlob(ctx context.Context, glob string) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if err := r.WalkConfigurationFilesWithGlob(
		ctx,
		"",
		glob,
		r.giterminismConfig.IsUncommittedConfigGoTemplateRenderingFileAccepted,
		func(relPath string, data []byte, err error) error {
			if err != nil {
				return err
			}

			result[filepath.ToSlash(relPath)] = string(data)

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("{{ .Files.Glob %q }}: %s", glob, err)
	}

	return result, nil
}

func (r FileReader) ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error) {
	data, err := r.ReadAndCheckConfigurationFile(ctx, relPath, r.giterminismConfig.IsUncommittedConfigGoTemplateRenderingFileAccepted)
	if err != nil {
		return nil, fmt.Errorf("{{ .Files.Get %q }}: %s", relPath, err)
	}

	return data, nil
}
