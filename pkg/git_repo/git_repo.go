package git_repo

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

// TODO: This is interface for build pkg only -- should be renamed.
// Do not add operations that are not designed for build pkg usage.
type GitRepo interface {
	String() string
	GetName() string

	IsEmpty() (bool, error)
	HeadCommit() (string, error)
	LatestBranchCommit(branch string) (string, error)
	LatestTagCommit(tag string) (string, error)
	FindCommitIdByMessage(regex string) (string, error)

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
