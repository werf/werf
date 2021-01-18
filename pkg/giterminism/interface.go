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
}

type Config interface {
	IsUncommittedConfigAccepted() bool
	IsUncommittedConfigTemplateFileAccepted(relPath string) (bool, error)
	IsUncommittedConfigGoTemplateRenderingFileAccepted(relPath string) (bool, error)
	IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error)
}
