package giterminism_manager

import (
	"context"

	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

type Interface interface {
	FileReader() FileReader
	Inspector() Inspector

	LocalGitRepo() git_repo.GitRepo
	HeadCommit() string
	ProjectDir() string
	RelativeToGitProjectDir() string
	LooseGiterminism() bool
	Dev() bool
}

type FileReader interface {
	IsConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadConfig(ctx context.Context, customRelPath string) (string, []byte, error)
	ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error
	ConfigGoTemplateFilesExists(ctx context.Context, relPath string) (bool, error)
	ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error)
	ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error)
	ConfigGoTemplateFilesIsDir(ctx context.Context, relPath string) (bool, error)
	ReadDockerfile(ctx context.Context, relPath string) ([]byte, error)
	IsDockerignoreExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadDockerignore(ctx context.Context, relPath string) ([]byte, error)

	file.ChartFileReader
}

type Inspector interface {
	InspectCustomTags() error
	InspectConfigGoTemplateRenderingEnv(ctx context.Context, envName string) error
	InspectConfigStapelFromLatest() error
	InspectConfigStapelGitBranch() error
	InspectConfigStapelMountBuildDir() error
	InspectConfigStapelMountFromPath(fromPath string) error
	InspectConfigDockerfileContextAddFile(relPath string) error
	InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error
	InspectConfigSecretEnvAccepted(secret string) error
	InspectConfigSecretSrcAccepted(secret string) error
	InspectConfigSecretValueAccepted(secret string) error
}
