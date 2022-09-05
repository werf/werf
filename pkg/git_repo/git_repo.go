package git_repo

import (
	"context"
	"path/filepath"
	"time"

	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const GitReposCacheVersion = "5"

type (
	PatchOptions    true_git.PatchOptions
	ArchiveOptions  true_git.ArchiveOptions
	LsTreeOptions   ls_tree.LsTreeOptions
	ChecksumOptions struct {
		LsTreeOptions
		Commit string
	}
)

func (opts ChecksumOptions) ID() string {
	return util.Sha256Hash(opts.Commit, ls_tree.LsTreeOptions(opts.LsTreeOptions).ID())
}

type ArchiveType string

const (
	FileArchive      ArchiveType = "file"
	DirectoryArchive ArchiveType = "directory"
)

type GitRepo interface {
	String() string
	GetName() string
	IsLocal() bool
	GetWorkTreeDir() string
	RemoteOriginUrl(_ context.Context) (string, error)
	IsShallowClone(ctx context.Context) (bool, error)
	FetchOrigin(ctx context.Context, opts FetchOptions) error
	Unshallow(ctx context.Context) error
	SyncWithOrigin(ctx context.Context) error

	CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error)
	GetCommitTreeEntry(ctx context.Context, commit, path string) (*ls_tree.LsTreeEntry, error)
	GetMergeCommitParents(ctx context.Context, commit string) ([]string, error)
	GetOrCreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error)
	GetOrCreateChecksum(ctx context.Context, opts ChecksumOptions) (string, error)
	GetOrCreatePatch(ctx context.Context, opts PatchOptions) (Patch, error)
	HeadCommitHash(ctx context.Context) (string, error)
	HeadCommitTime(ctx context.Context) (*time.Time, error)
	IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error)
	IsCommitDirectoryExist(ctx context.Context, commit, path string) (bool, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
	IsCommitFileExist(ctx context.Context, commit, path string) (bool, error)
	IsAnyCommitTreeEntriesMatched(ctx context.Context, commit, pathScope string, pathMatcher path_matcher.PathMatcher, allFiles bool) (bool, error)
	IsCommitTreeEntryDirectory(ctx context.Context, commit, relPath string) (bool, error)
	IsCommitTreeEntryExist(ctx context.Context, commit, relPath string) (bool, error)
	IsEmpty(ctx context.Context) (bool, error)
	LatestBranchCommit(ctx context.Context, branch string) (string, error)
	ListCommitFilesWithGlob(ctx context.Context, commit, dir, glob string) ([]string, error)
	ReadCommitFile(ctx context.Context, commit, path string) ([]byte, error)
	ReadCommitTreeEntryContent(ctx context.Context, commit, relPath string) ([]byte, error)
	ResolveAndCheckCommitFilePath(ctx context.Context, commit, path string, checkSymlinkTargetFunc func(resolvedPath string) error) (string, error)
	ResolveCommitFilePath(ctx context.Context, commit, path string) (string, error)
	TagCommit(ctx context.Context, tag string) (string, error)
	WalkCommitFiles(ctx context.Context, commit, dir string, pathMatcher path_matcher.PathMatcher, fileFunc func(notResolvedPath string) error) error

	StatusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher) (list []string, err error)
	ValidateStatusResult(ctx context.Context, pathMatcher path_matcher.PathMatcher) error
}

type FetchOptions struct {
	Unshallow bool
}

type gitRepo interface {
	GitRepo

	withRepoHandle(ctx context.Context, commit string, f func(handle repo_handle.Handle) error) error
}

type Patch interface {
	GetFilePath() string
	IsEmpty() bool
	HasBinary() bool
	GetPaths() []string
	GetBinaryPaths() []string
	GetPathsToRemove() []string
}

type Archive interface {
	GetFilePath() string
}

func GetGitRepoCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_repos", GitReposCacheVersion)
}
