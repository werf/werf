package file_reader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/util"
)

func (r FileReader) LocateChart(ctx context.Context, relDir string, _ *cli.EnvSettings) (string, error) {
	files, err := r.loadChartDir(ctx, relDir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", NewFilesNotFoundInTheProjectGitRepositoryError(chartDirectoryErrorConfigType, relDir)
	}

	return relDir, nil
}

func (r FileReader) ReadChartFile(ctx context.Context, relPath string) ([]byte, error) {
	if err := r.checkChartFileExistence(ctx, relPath); err != nil {
		return nil, err
	}

	return r.readChartFile(ctx, relPath)
}

func (r FileReader) LoadChartDir(ctx context.Context, relDir string) ([]*chart.ChartExtenderBufferedFile, error) {
	return r.loadChartDir(ctx, relDir)
}

// TODO helmignore support
func (r FileReader) loadChartDir(ctx context.Context, relDir string) ([]*chart.ChartExtenderBufferedFile, error) {
	var res []*chart.ChartExtenderBufferedFile

	if exist, err := r.isConfigurationFileExistAnywhere(ctx, relDir); err != nil {
		return nil, err
	} else if !exist {
		return nil, NewFilesNotFoundInTheProjectGitRepositoryError(chartDirectoryErrorConfigType, relDir)
	}

	// TODO configurationFilesGlob method must resolve symlinks properly
	relDir, err := r.resolveChartDirectory(relDir)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(relDir, "**/*")
	if err := r.configurationFilesGlob(
		ctx,
		chartFileErrorConfigType,
		pattern,
		r.giterminismConfig.IsUncommittedHelmFileAccepted,
		r.readChartFile,
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

func (r FileReader) resolveChartDirectory(relDir string) (string, error) {
	absDir := filepath.Join(r.sharedOptions.ProjectDir(), relDir)
	link, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		return "", fmt.Errorf("eval symlink %s failed: %s", absDir, err)
	}

	linkStat, err := os.Lstat(link)
	if err != nil {
		return "", fmt.Errorf("lstat %s failed: %s", linkStat, err)
	}

	if !linkStat.IsDir() {
		return "", fmt.Errorf("unable to handle the chart directory '%s': linked to file not a directory", link)
	}

	if !util.IsSubpathOfBasePath(r.sharedOptions.ProjectDir(), link) {
		return "", fmt.Errorf("unable to handle the chart directory '%s' which is located outside the project directory", link)
	}

	return util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), link), nil
}

func (r FileReader) readChartFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.readConfigurationFile(ctx, chartFileErrorConfigType, relPath, r.giterminismConfig.IsUncommittedHelmFileAccepted)
}

func (r FileReader) checkChartFileExistence(ctx context.Context, relPath string) error {
	return r.checkConfigurationFileExistence(ctx, chartFileErrorConfigType, relPath, r.giterminismConfig.IsUncommittedHelmFileAccepted)
}
