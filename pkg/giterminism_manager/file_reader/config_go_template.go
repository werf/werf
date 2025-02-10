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
		r.giterminismConfig.UncommittedConfigGoTemplateRenderingFilePathMatcher(),
		func(relativeToDirNotResolvedPath string, data []byte, err error) error {
			if err != nil {
				return err
			}

			result[filepath.ToSlash(relativeToDirNotResolvedPath)] = string(data)

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("{{ .Files.Glob %q }}: %w", glob, err)
	}

	return result, nil
}

func (r FileReader) ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error) {
	data, err := r.ReadAndCheckConfigurationFile(ctx, relPath,
		r.giterminismConfig.UncommittedConfigGoTemplateRenderingFilePathMatcher().IsPathMatched,
		func(path string) (bool, error) {
			return r.IsRegularFileExist(ctx, path)
		})
	if err != nil {
		return nil, fmt.Errorf("{{ .Files.Get %q }}: %w", relPath, err)
	}

	return data, nil
}

func (r FileReader) ConfigGoTemplateFilesExists(ctx context.Context, relPath string) (bool, error) {
	err := r.CheckFileExistenceAndAcceptance(ctx, relPath,
		r.giterminismConfig.UncommittedConfigGoTemplateRenderingFilePathMatcher().IsPathMatched,
		func(path string) (bool, error) {
			return r.IsFileExist(ctx, path)
		})
	if err != nil {
		return false, fmt.Errorf("{{ .Files.Exists %q }}: %w", relPath, err)
	}

	return r.IsFileExist(ctx, relPath)
}

func (r FileReader) ConfigGoTemplateFilesIsDir(ctx context.Context, relPath string) (bool, error) {
	err := r.CheckFileExistenceAndAcceptance(ctx, relPath,
		r.giterminismConfig.UncommittedConfigGoTemplateRenderingFilePathMatcher().IsPathMatched,
		func(path string) (bool, error) {
			return r.IsDirectoryExist(ctx, path)
		})
	if err != nil {
		return false, fmt.Errorf("{{ .Files.IsDir %q }}: %w", relPath, err)
	}
	return r.IsDirectoryExist(ctx, relPath)
}
