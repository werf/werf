package giterminism_manager

import (
	"context"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/path_matcher"
)

type Interface interface {
	FileReader() FileReader
	Inspector() Inspector

	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	ProjectDir() string
	Dev() bool
}

type FileReader interface {
	IsConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadConfig(ctx context.Context, customRelPath string) ([]byte, error)
	ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error
	ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error)
	ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error)
	ReadDockerfile(ctx context.Context, relPath string) ([]byte, error)
	IsDockerignoreExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadDockerignore(ctx context.Context, relPath string) ([]byte, error)

	HelmChartExtender
}

type HelmChartExtender interface {
	LocateChart(ctx context.Context, name string, settings *cli.EnvSettings) (string, error)
	ReadChartFile(ctx context.Context, filePath string) ([]byte, error)
	LoadChartDir(ctx context.Context, dir string) ([]*chart.ChartExtenderBufferedFile, error)
}

type Inspector interface {
	InspectConfigGoTemplateRenderingEnv(ctx context.Context, envName string) error
	InspectConfigStapelFromLatest() error
	InspectConfigStapelGitBranch() error
	InspectConfigStapelMountBuildDir() error
	InspectConfigStapelMountFromPath(fromPath string) error
	InspectConfigDockerfileContextAddFile(relPath string) error
	InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error
}
