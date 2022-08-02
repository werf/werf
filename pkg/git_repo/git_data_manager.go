package git_repo

import (
	"context"

	"github.com/werf/lockgate"
	"github.com/werf/werf/pkg/true_git"
)

type GitDataManager interface {
	CreateArchiveFile(ctx context.Context, repoID string, opts ArchiveOptions, tmpPath string) (*ArchiveFile, error)
	GetArchiveFile(ctx context.Context, repoID string, opts ArchiveOptions) (*ArchiveFile, error)
	CreatePatchFile(ctx context.Context, repoID string, opts PatchOptions, tmpPath string, desc *true_git.PatchDescriptor) (*PatchFile, error)
	GetPatchFile(ctx context.Context, repoID string, opts PatchOptions) (*PatchFile, error)
	NewTmpFile() (string, error)
	LockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error)

	GetArchivesCacheDir() string
	GetPatchesCacheDir() string
}
