package git_repo

import (
	"io"
)

type FilterOptions struct {
	BasePath                   string
	IncludePaths, ExcludePaths []string
}

type PatchOptions struct {
	FilterOptions
	FromCommit, ToCommit string
}

type ArchiveOptions struct {
	FilterOptions
	Commit string
}

type ArchiveType string

const (
	FileArchive      ArchiveType = "file"
	DirectoryArchive ArchiveType = "directory"
)

type Patch interface {
	GetFilePath() string
	IsEmpty() (bool, error)
}

type GitRepo interface {
	String() string

	HeadCommit() (string, error)
	HeadBranchName() (string, error)
	LatestBranchCommit(branch string) (string, error)
	LatestTagCommit(tag string) (string, error)

	CreatePatch(PatchOptions) (Patch, error)

	// TODO: change main interface to CreateArchive (Archive, error)
	ArchiveType(ArchiveOptions) (ArchiveType, error)
	IsAnyEntries(ArchiveOptions) (bool, error)
	CreateArchiveTar(io.Writer, ArchiveOptions) error
	ArchiveChecksum(ArchiveOptions) (string, error) // TODO
}
