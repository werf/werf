package file_reader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/export/helm/werf/file"
	"github.com/werf/nelm/pkg/helm/pkg/ignore"
)

func (r FileReader) LocateChart(ctx context.Context, chartDir string) (string, error) {
	chartDir, err := r.locateChart(ctx, chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to locate chart directory: %w", err)
	}

	return chartDir, nil
}

func (r FileReader) locateChart(ctx context.Context, chartDir string) (string, error) {
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
	return r.ReadAndCheckConfigurationFile(ctx, relPath, r.giterminismConfig.UncommittedHelmFilePathMatcher().IsPathMatched,
		func(path string) (bool, error) {
			return r.IsRegularFileExist(ctx, path)
		})
}

func (r FileReader) LoadChartDir(ctx context.Context, chartDir string) ([]*file.ChartExtenderBufferedFile, error) {
	relDir := r.absolutePathToProjectDirRelativePath(chartDir)

	files, err := r.loadChartDir(ctx, relDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory: %w", err)
	}

	return files, nil
}

func (r FileReader) loadChartDir(ctx context.Context, relDir string) ([]*file.ChartExtenderBufferedFile, error) {
	rules, err := r.readChartIgnoreRules(ctx, relDir)
	if err != nil {
		return nil, err
	}

	var res []*file.ChartExtenderBufferedFile

	if err := r.WalkConfigurationFilesWithGlob(
		ctx,
		relDir,
		"**/*",
		r.giterminismConfig.UncommittedHelmFilePathMatcher(),
		func(relativeToDirNotResolvedPath string, data []byte, err error) error {
			relativeToDirNotResolvedPath = filepath.ToSlash(relativeToDirNotResolvedPath)

			// An ignored file is not a part of the chart, so its git state is irrelevant:
			// the check must precede err to let users exclude uncommitted files.
			if matchChartIgnoreRules(rules, relativeToDirNotResolvedPath) {
				return nil
			}

			if err != nil {
				return err
			}

			res = append(res, &file.ChartExtenderBufferedFile{Name: relativeToDirNotResolvedPath, Data: data})

			return nil
		},
	); err != nil {
		return nil, err
	}

	return res, nil
}

func (r FileReader) readChartIgnoreRules(ctx context.Context, relDir string) (*ignore.Rules, error) {
	relPath := filepath.Join(relDir, ignore.HelmIgnore)

	exist, err := r.IsConfigurationFileExistAnywhere(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("check %q existence: %w", filepath.ToSlash(relPath), err)
	}

	var data []byte
	if exist {
		data, err = r.readChartFile(ctx, relPath)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", filepath.ToSlash(relPath), err)
		}
	}

	rules, err := parseChartIgnoreRules(data)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", filepath.ToSlash(relPath), err)
	}

	return rules, nil
}

// parseChartIgnoreRules builds the rule set for a chart directory, where nil data means
// the chart has no .helmignore. Helm applies its default rules either way.
func parseChartIgnoreRules(data []byte) (*ignore.Rules, error) {
	rules := ignore.Empty()
	if data != nil {
		var err error
		rules, err = ignore.Parse(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}
	rules.AddDefaults()

	return rules, nil
}

// matchChartIgnoreRules reports whether the chart-relative file path is ignored.
// Helm drops an ignored directory by returning filepath.SkipDir while walking, so its rules
// match the directory itself rather than the files under it. This walk yields files only,
// hence every ancestor directory is matched explicitly to keep the same semantics.
func matchChartIgnoreRules(rules *ignore.Rules, relPath string) bool {
	if rules.Ignore(relPath, chartIgnoreFileInfo{}) {
		return true
	}

	parts := strings.Split(relPath, "/")
	for i := 1; i < len(parts); i++ {
		if rules.Ignore(strings.Join(parts[:i], "/"), chartIgnoreFileInfo{isDir: true}) {
			return true
		}
	}

	return false
}

// chartIgnoreFileInfo carries the only bit ignore.Rules reads from os.FileInfo.
type chartIgnoreFileInfo struct {
	isDir bool
}

var _ os.FileInfo = chartIgnoreFileInfo{}

func (i chartIgnoreFileInfo) Name() string       { return "" }
func (i chartIgnoreFileInfo) Size() int64        { return 0 }
func (i chartIgnoreFileInfo) Mode() os.FileMode  { return 0 }
func (i chartIgnoreFileInfo) ModTime() time.Time { return time.Time{} }
func (i chartIgnoreFileInfo) IsDir() bool        { return i.isDir }
func (i chartIgnoreFileInfo) Sys() any           { return nil }

// This method exists only for backward compatibility for Loader interface
func (r FileReader) ChartIsDir(relPath string) (bool, error) {
	absPath := util.GetAbsoluteFilepath(relPath)

	fi, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("path %q not found on local filesystem", relPath)
		}
		return false, fmt.Errorf("os.Stat failed for %q: %w", absPath, err)
	}
	return fi.IsDir(), nil
}
