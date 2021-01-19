package giterminism

import (
	"context"

	"helm.sh/helm/v3/pkg/chart"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/git_repo"
)

type Manager interface {
	FileReader() FileReader
	Inspector() Inspector
	Config() Config

	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	ProjectDir() string

	LooseGiterminism() bool
}

type FileReader interface {
	IsGiterminismConfigExistAnywhere(ctx context.Context) (bool, error)
	ReadGiterminismConfig(ctx context.Context) ([]byte, error)
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
}

type Config interface {
	IsUncommittedConfigAccepted() bool
	IsUncommittedConfigTemplateFileAccepted(relPath string) (bool, error)
	IsUncommittedConfigGoTemplateRenderingFileAccepted(relPath string) (bool, error)
	IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error)
	IsConfigStapelFromLatestAccepted() bool
	IsConfigStapelGitBranchAccepted() bool
	IsConfigStapelMountBuildDirAccepted() bool
	IsConfigStapelMountFromPathAccepted(fromPath string) (bool, error)
	IsConfigDockerfileContextAddFileAccepted(relPath string) (bool, error)
	IsUncommittedDockerfileAccepted(relPath string) (bool, error)
	IsUncommittedDockerignoreAccepted(relPath string) (bool, error)
	IsUncommittedHelmFileAccepted(relPath string) (bool, error)
}
