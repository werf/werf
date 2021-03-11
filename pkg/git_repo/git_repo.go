package git_repo

import (
	"context"
	"path/filepath"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const GitRepoCacheVersion = "4"

type PatchOptions true_git.PatchOptions
type ArchiveOptions true_git.ArchiveOptions
type ChecksumOptions struct {
	PathMatcher path_matcher.PathMatcher
	Commit      string
}

func (opts ChecksumOptions) ID() string {
	return util.Sha256Hash(opts.Commit, opts.PathMatcher.ID())
}

type ArchiveType string

const (
	FileArchive      ArchiveType = "file"
	DirectoryArchive ArchiveType = "directory"
)

type GitRepo interface {
	String() string
	GetName() string

	IsEmpty(ctx context.Context) (bool, error)
	HeadCommit(ctx context.Context) (string, error)
	LatestBranchCommit(ctx context.Context, branch string) (string, error)
	TagCommit(ctx context.Context, tag string) (string, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
	IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error)

	GetMergeCommitParents(ctx context.Context, commit string) ([]string, error)

	CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error)
	GetOrCreatePatch(context.Context, PatchOptions) (Patch, error)
	GetOrCreateArchive(context.Context, ArchiveOptions) (Archive, error)
	GetOrCreateChecksum(context.Context, ChecksumOptions) (string, error)
}

type Patch interface {
	GetFilePath() string
	IsEmpty() bool
	HasBinary() bool
	GetPaths() []string
	GetBinaryPaths() []string
}

type Archive interface {
	GetFilePath() string
	GetType() ArchiveType
	IsEmpty() bool
}

func GetGitRepoCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_repos", GitRepoCacheVersion)
}
