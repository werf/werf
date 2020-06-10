package git_repo

import (
	"path/filepath"

	"github.com/werf/werf/pkg/werf"
)

const GitRepoCacheVersion = "3"

type PatchOptions struct {
	FilterOptions
	FromCommit, ToCommit string

	WithEntireFileContext bool
	WithBinary            bool
}

type ArchiveOptions struct {
	FilterOptions
	Commit string
}

type ChecksumOptions struct {
	FilterOptions
	Paths  []string
	Commit string
}

type FilterOptions struct {
	BasePath                   string
	IncludePaths, ExcludePaths []string
}

type ArchiveType string

const (
	FileArchive      ArchiveType = "file"
	DirectoryArchive ArchiveType = "directory"
)

type GitRepo interface {
	String() string
	GetName() string

	IsEmpty() (bool, error)
	HeadCommit() (string, error)
	LatestBranchCommit(branch string) (string, error)
	TagCommit(tag string) (string, error)
	IsCommitExists(commit string) (bool, error)
	IsAncestor(ancestorCommit, descendantCommit string) (bool, error)

	GetMergeCommitParents(commit string) ([]string, error)

	CreateDetachedMergeCommit(fromCommit, toCommit string) (string, error)
	CreatePatch(PatchOptions) (Patch, error)
	CreateArchive(ArchiveOptions) (Archive, error)
	Checksum(ChecksumOptions) (Checksum, error)
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

type Checksum interface {
	String() string
	GetNoMatchPaths() []string
}

func GetGitRepoCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_repos", GitRepoCacheVersion)
}
