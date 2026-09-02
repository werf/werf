package file_reader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/common-go/pkg/util"
	nelmcommon "github.com/werf/nelm/pkg/common"
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

	rules, err := r.readChartIgnoreRules(ctx, relDir, ReadChartIgnoreRulesOptions{})
	if err != nil {
		return "", err
	}

	files, err := r.loadChartDir(ctx, relDir, rules)
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

func (r FileReader) ChartFileExists(ctx context.Context, path string) (bool, error) {
	relPath := r.absolutePathToProjectDirRelativePath(path)
	return r.IsRegularFileExist(ctx, relPath)
}

func (r FileReader) readChartFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, relPath, r.giterminismConfig.UncommittedHelmFilePathMatcher().IsPathMatched,
		func(path string) (bool, error) {
			return r.IsRegularFileExist(ctx, path)
		})
}

func (r FileReader) LoadChartDir(ctx context.Context, chartDir string) ([]*nelmcommon.BufferedFile, error) {
	relDir := r.absolutePathToProjectDirRelativePath(chartDir)

	rules, err := r.readChartIgnoreRules(ctx, relDir, ReadChartIgnoreRulesOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory: %w", err)
	}

	files, err := r.loadChartDir(ctx, relDir, rules)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory: %w", err)
	}

	return files, nil
}

// LoadChartDirWithIgnoreRules loads the chart directory excluding the files matched by the passed
// rules. It exists because the effective chart tree can be assembled from more than the local
// directory, in which case the winning .helmignore is resolved by the caller and has to be applied
// here too — while the file is still excluded before it gets read or checked against giterminism.
func (r FileReader) LoadChartDirWithIgnoreRules(ctx context.Context, chartDir string, rules ChartIgnoreRules) ([]*nelmcommon.BufferedFile, error) {
	relDir := r.absolutePathToProjectDirRelativePath(chartDir)

	files, err := r.loadChartDir(ctx, relDir, rules)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory: %w", err)
	}

	return files, nil
}

func (r FileReader) loadChartDir(ctx context.Context, relDir string, rules ChartIgnoreRules) ([]*nelmcommon.BufferedFile, error) {
	var res []*nelmcommon.BufferedFile

	if err := r.WalkConfigurationFilesWithGlob(
		ctx,
		relDir,
		"**/*",
		r.giterminismConfig.UncommittedHelmFilePathMatcher(),
		func(relativeToDirNotResolvedPath string, data []byte, err error) error {
			if err != nil {
				return err
			}

			res = append(res, &nelmcommon.BufferedFile{Name: filepath.ToSlash(relativeToDirNotResolvedPath), Data: data})

			return nil
		},
		WalkConfigurationFilesWithGlobOptions{
			// An ignored file is not a part of the chart, so it is excluded before it gets read
			// or checked against the giterminism config: its git state is irrelevant, and
			// .helmignore stays usable to exclude uncommitted files.
			SkipRelativeToDirPathFunc: func(relativeToDirPath string, isDir bool) bool {
				return matchChartIgnoreRules(rules.rules, relativeToDirPath, isDir)
			},
		},
	); err != nil {
		return nil, err
	}

	return res, nil
}

// ChartIgnoreRules is the .helmignore rule set applied to a chart directory. Its zero value
// excludes nothing, which is what a chart without a .helmignore needs before helm's own defaults
// are added on top.
type ChartIgnoreRules struct {
	rules         *ignore.Rules
	hasIgnoreFile bool
}

// HasIgnoreFile reports whether the rules came from an actual .helmignore — the chart's own or one
// supplied by the caller — rather than from helm's defaults alone. An empty .helmignore counts as
// one, so this cannot be inferred from the rule set.
func (r ChartIgnoreRules) HasIgnoreFile() bool {
	return r.hasIgnoreFile
}

// IsFileIgnored reports whether a chart-relative file path is excluded, either by a rule matching
// the file itself or by a rule matching one of its parent directories. Unlike a filesystem walk, a
// flat file list cannot have a directory pruned, so the parents have to be checked explicitly.
func (r ChartIgnoreRules) IsFileIgnored(ctx context.Context, relPath string) bool {
	return skipRelativeToDirPathWithParents(filepath.ToSlash(relPath), func(path string, isDir bool) bool {
		return matchChartIgnoreRules(r.rules, path, isDir)
	})
}

type ReadChartIgnoreRulesOptions struct {
	// FallbackData is the .helmignore content to use when the chart directory has no local one.
	FallbackData []byte
	// FallbackExists is separate from FallbackData because an empty .helmignore yields no data
	// while still being present.
	FallbackExists bool
}

// ReadChartIgnoreRules resolves the .helmignore rule set of a chart directory. The local file wins
// over the fallback, which lets a caller that assembles the chart from several sources keep the
// documented precedence of local project files.
func (r FileReader) ReadChartIgnoreRules(ctx context.Context, chartDir string, opts ReadChartIgnoreRulesOptions) (ChartIgnoreRules, error) {
	return r.readChartIgnoreRules(ctx, r.absolutePathToProjectDirRelativePath(chartDir), opts)
}

func (r FileReader) readChartIgnoreRules(ctx context.Context, relDir string, opts ReadChartIgnoreRulesOptions) (ChartIgnoreRules, error) {
	relPath := filepath.Join(relDir, ignore.HelmIgnore)

	exist, err := r.IsConfigurationFileExistAnywhere(ctx, relPath)
	if err != nil {
		return ChartIgnoreRules{}, fmt.Errorf("check %q existence: %w", filepath.ToSlash(relPath), err)
	}

	data := opts.FallbackData
	hasIgnoreFile := opts.FallbackExists
	if exist {
		hasIgnoreFile = true

		data, err = r.readChartFile(ctx, relPath)
		if err != nil {
			return ChartIgnoreRules{}, fmt.Errorf("read %q: %w", filepath.ToSlash(relPath), err)
		}
	}

	rules, err := parseChartIgnoreRules(data, hasIgnoreFile)
	if err != nil {
		return ChartIgnoreRules{}, fmt.Errorf("parse %q: %w", filepath.ToSlash(relPath), err)
	}

	return rules, nil
}

// newChartIgnoreRules is the only place helm's defaults are added, so a rule set built from a
// parsed .helmignore and one built without a file cannot drift apart.
func newChartIgnoreRules(rules *ignore.Rules, hasIgnoreFile bool) ChartIgnoreRules {
	rules.AddDefaults()

	return ChartIgnoreRules{rules: rules, hasIgnoreFile: hasIgnoreFile}
}

func parseChartIgnoreRules(data []byte, hasIgnoreFile bool) (ChartIgnoreRules, error) {
	if data == nil {
		return newChartIgnoreRules(ignore.Empty(), hasIgnoreFile), nil
	}

	rules, err := ignore.Parse(bytes.NewReader(data))
	if err != nil {
		return ChartIgnoreRules{}, err
	}

	return newChartIgnoreRules(rules, hasIgnoreFile), nil
}

// matchChartIgnoreRules reports whether the chart-relative path is ignored. A matched directory
// is skipped along with its whole subtree, which is how helm applies directory rules.
func matchChartIgnoreRules(rules *ignore.Rules, relPath string, isDir bool) bool {
	if rules == nil {
		return false
	}

	return rules.Ignore(relPath, chartIgnoreFileInfo{path: relPath, isDir: isDir})
}

type chartIgnoreFileInfo struct {
	path  string
	isDir bool
}

var _ os.FileInfo = chartIgnoreFileInfo{}

func (i chartIgnoreFileInfo) Name() string       { return filepath.Base(i.path) }
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
