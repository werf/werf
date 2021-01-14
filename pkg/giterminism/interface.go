package giterminism

import (
	"context"

	"github.com/werf/werf/pkg/git_repo"
)

type Manager interface {
	FileReader() FileReader
	Config() Config

	LocalGitRepo() *git_repo.Local
	HeadCommit() string
	ProjectDir() string

	LooseGiterminism() bool
}

type FileReader interface {
	ReadConfig(ctx context.Context, customRelPath string) ([]byte, error)
	ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(relPath string, data []byte, err error) error) error
}

type Config interface {
	IsUncommittedConfigAccepted() bool
	IsUncommittedConfigTemplateFileAccepted(relPath string) (bool, error)
}
