package container_backend

import "context"

type BuildContextArchiver interface {
	Create(ctx context.Context, opts BuildContextArchiveCreateOptions) error
	Path() string
	ExtractOrGetExtractedDir(ctx context.Context) (string, error)
	CleanupExtractedDir(ctx context.Context)
}

type BuildContextArchiveCreateOptions struct {
	DockerfileRelToContextPath string
	ContextGitSubDir           string
	ContextAddFiles            []string
}
