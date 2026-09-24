package container_backend

import "context"

type BuildContextArchiver interface {
	Create(ctx context.Context, opts BuildContextArchiveCreateOptions) error
	Path() string
	ExtractOrGetExtractedDir(ctx context.Context) (string, error)
	CalculatePathsChecksum(ctx context.Context, paths []string) (string, error)
	CalculateGlobsChecksum(ctx context.Context, globs []string, opts CalculateGlobsChecksumOptions) (string, error)
	CleanupExtractedDir(ctx context.Context)
}

type CalculateGlobsChecksumOptions struct {
	CheckForArchives bool
	// COPY --parents turns matched paths into destination paths, so their names affect the result
	IncludeMatchedPaths bool
}

type BuildContextArchiveCreateOptions struct {
	DockerfileRelToContextPath string
	ContextGitSubDir           string
	ContextAddFiles            []string
}
