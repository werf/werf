package giterminism

import (
	"context"

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
	ReadConfig(ctx context.Context, customRelPath string) ([]byte, error)
	ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error
	ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error)
	ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error)
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
}
