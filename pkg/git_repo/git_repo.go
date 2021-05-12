package git_repo

import (
	"context"
	"path/filepath"

	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const GitReposCacheVersion = "5"

type PatchOptions true_git.PatchOptions
type ArchiveOptions true_git.ArchiveOptions
type LsTreeOptions ls_tree.LsTreeOptions
type ChecksumOptions struct {
	LsTreeOptions
	Commit string
}

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

	CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error)
	GetCommitTreeEntry(ctx context.Context, commit string, path string) (*ls_tree.LsTreeEntry, error)
	GetMergeCommitParents(ctx context.Context, commit string) ([]string, error)
	GetOrCreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error)
	GetOrCreateChecksum(ctx context.Context, opts ChecksumOptions) (string, error)
	GetOrCreatePatch(ctx context.Context, opts PatchOptions) (Patch, error)
	HeadCommit(ctx context.Context) (string, error)
	IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error)
	IsCommitDirectoryExist(ctx context.Context, commit, path string) (bool, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
	IsCommitFileExist(ctx context.Context, commit, path string) (bool, error)
	IsNoCommitTreeEntriesMatched(ctx context.Context, commit string, pathScope string, pathMatcher path_matcher.PathMatcher) (bool, error)
	IsCommitTreeEntryDirectory(ctx context.Context, commit string, relPath string) (bool, error)
	IsCommitTreeEntryExist(ctx context.Context, commit string, relPath string) (bool, error)
	IsEmpty(ctx context.Context) (bool, error)
	LatestBranchCommit(ctx context.Context, branch string) (string, error)
	ListCommitFilesWithGlob(ctx context.Context, commit string, dir string, glob string) ([]string, error)
	ReadCommitFile(ctx context.Context, commit, path string) ([]byte, error)
	ReadCommitTreeEntryContent(ctx context.Context, commit string, relPath string) ([]byte, error)
	ResolveAndCheckCommitFilePath(ctx context.Context, commit, path string, checkSymlinkTargetFunc func(resolvedPath string) error) (string, error)
	ResolveCommitFilePath(ctx context.Context, commit, path string) (string, error)
	TagCommit(ctx context.Context, tag string) (string, error)
	WalkCommitFiles(ctx context.Context, commit string, dir string, pathMatcher path_matcher.PathMatcher, fileFunc func(notResolvedPath string) error) error

	initRepoHandleBackedByWorkTree(ctx context.Context, commit string) (repo_handle.Handle, error)
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
	IsEmpty() bool
}

func GetGitRepoCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_repos", GitReposCacheVersion)
}
