package file_reader

import (
	"context"
	"fmt"
	"path/filepath"
)

func (r FileReader) ReadDockerfile(ctx context.Context, relPath string) ([]byte, error) {
	data, err := r.readDockerfile(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dockerfile %q: %s", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (r FileReader) readDockerfile(ctx context.Context, relPath string) ([]byte, error) {
	if err := r.checkConfigurationFileExistence(ctx, relPath, r.giterminismConfig.IsUncommittedDockerfileAccepted); err != nil {
		return nil, err
	}

	return r.readConfigurationFile(ctx, relPath, r.giterminismConfig.IsUncommittedDockerfileAccepted)
}

func (r FileReader) IsDockerignoreExistAnywhere(ctx context.Context, relPath string) (bool, error) {
	return r.isConfigurationFileExistAnywhere(ctx, relPath)
}

func (r FileReader) ReadDockerignore(ctx context.Context, relPath string) ([]byte, error) {
	data, err := r.readDockerignore(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dockerignore file %q: %s", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (r FileReader) readDockerignore(ctx context.Context, relPath string) ([]byte, error) {
	if err := r.checkConfigurationFileExistence(ctx, relPath, r.giterminismConfig.IsUncommittedDockerignoreAccepted); err != nil {
		return nil, err
	}

	return r.readConfigurationFile(ctx, relPath, r.giterminismConfig.IsUncommittedDockerignoreAccepted)
}

func (r FileReader) isDockerignoreExist(ctx context.Context, relPath string) (bool, error) {
	return r.isConfigurationFileExist(ctx, relPath, r.giterminismConfig.IsUncommittedDockerignoreAccepted)
}
