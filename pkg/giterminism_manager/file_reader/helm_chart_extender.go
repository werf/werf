package file_reader

import (
	"context"
	"fmt"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
)

func (r FileReader) LocateChart(ctx context.Context, chartDir string, settings *cli.EnvSettings) (string, error) {
	chartDir, err := r.locateChart(ctx, chartDir, settings)
	if err != nil {
		return "", fmt.Errorf("unable to locate chart directory: %w", err)
	}

	return chartDir, nil
}

func (r FileReader) locateChart(ctx context.Context, chartDir string, _ *cli.EnvSettings) (string, error) {
	relDir := r.absolutePathToProjectDirRelativePath(chartDir)

	files, err := r.loadChartDir(ctx, relDir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("the directory %q not found in the project git repository", relDir)
	}

	return chartDir, nil
}

func (r FileReader) ReadChartFile(ctx context.Context, path string) ([]byte, error) {
	relPath := r.absolutePathToProjectDirRelativePath(path)

	data, err := r.readChartFile(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read chart file %q: %w", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (r FileReader) readChartFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, relPath, r.giterminismConfig.UncommittedHelmFilePathMatcher().IsPathMatched)
}

func (r FileReader) LoadChartDir(ctx context.Context, chartDir string) ([]*chart.ChartExtenderBufferedFile, error) {
	relDir := r.absolutePathToProjectDirRelativePath(chartDir)

	files, err := r.loadChartDir(ctx, relDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory: %w", err)
	}

	return files, nil
}

// TODO helmignore support
func (r FileReader) loadChartDir(ctx context.Context, relDir string) ([]*chart.ChartExtenderBufferedFile, error) {
	var res []*chart.ChartExtenderBufferedFile

	if err := r.WalkConfigurationFilesWithGlob(
		ctx,
		relDir,
		"**/*",
		r.giterminismConfig.UncommittedHelmFilePathMatcher(),
		func(relativeToDirNotResolvedPath string, data []byte, err error) error {
			if err != nil {
				return err
			}

			relativeToDirNotResolvedPath = filepath.ToSlash(relativeToDirNotResolvedPath)
			res = append(res, &chart.ChartExtenderBufferedFile{Name: relativeToDirNotResolvedPath, Data: data})

			return nil
		},
	); err != nil {
		return nil, err
	}

	return res, nil
}
