package file_reader

import (
	"context"
	"fmt"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/util"
)

func (r FileReader) LocateChart(ctx context.Context, chartDir string, settings *cli.EnvSettings) (string, error) {
	chartDir, err := r.locateChart(ctx, chartDir, settings)
	if err != nil {
		return "", fmt.Errorf("unable to locate chart directory: %s", err)
	}

	return chartDir, nil
}

func (r FileReader) locateChart(ctx context.Context, chartDir string, _ *cli.EnvSettings) (string, error) {
	relDir := util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), chartDir)

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
	relPath := util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), path)

	data, err := r.readChartFile(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read chart file %q: %s", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (r FileReader) readChartFile(ctx context.Context, relPath string) ([]byte, error) {
	if err := r.CheckConfigurationFileExistence(ctx, relPath, r.giterminismConfig.IsUncommittedHelmFileAccepted); err != nil {
		return nil, err
	}

	return r.ReadAndValidateConfigurationFile(ctx, relPath, r.giterminismConfig.IsUncommittedHelmFileAccepted)
}

func (r FileReader) LoadChartDir(ctx context.Context, chartDir string) ([]*chart.ChartExtenderBufferedFile, error) {
	relDir := util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), chartDir)

	files, err := r.loadChartDir(ctx, relDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory: %s", err)
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
		r.giterminismConfig.IsUncommittedHelmFileAccepted,
		func(relPath string, data []byte, err error) error {
			if err != nil {
				return err
			}

			relPath = filepath.ToSlash(util.GetRelativeToBaseFilepath(relDir, relPath))
			res = append(res, &chart.ChartExtenderBufferedFile{Name: relPath, Data: data})

			return nil
		},
	); err != nil {
		return nil, err
	}

	return res, nil
}
