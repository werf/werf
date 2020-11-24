package git_repo

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
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

	IsEmpty(ctx context.Context) (bool, error)
	HeadCommit(ctx context.Context) (string, error)
	LatestBranchCommit(ctx context.Context, branch string) (string, error)
	TagCommit(ctx context.Context, tag string) (string, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
	IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error)

	GetMergeCommitParents(ctx context.Context, commit string) ([]string, error)

	CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error)
	CreatePatch(context.Context, PatchOptions) (Patch, error)
	CreateArchive(context.Context, ArchiveOptions) (Archive, error)
	Checksum(context.Context, ChecksumOptions) (Checksum, error)
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

func GetFileDataFromGitAndCompareWithLocal(localGitRepo *Local, commit, projectDir, relPath string) ([]byte, bool, error) {
	repoData, err := localGitRepo.ReadFile(commit, relPath)
	if err != nil {
		return nil, false, fmt.Errorf("unable to read file %s in local git repository: %s", relPath, err)
	}

	var localData []byte
	absPath := filepath.Join(projectDir, relPath)
	exist, err := util.FileExists(absPath)
	if err != nil {
		return nil, false, fmt.Errorf("unable to check file existance: %s", err)
	} else if exist {
		localData, err = ioutil.ReadFile(absPath)
		if err != nil {
			return nil, false, fmt.Errorf("unable to read file: %s", err)
		}
	}

	isDataIdentical := bytes.Equal(repoData, localData)
	localDataWithForcedUnixLineBreak := bytes.ReplaceAll(localData, []byte("\r\n"), []byte("\n"))
	if !isDataIdentical {
		isDataIdentical = bytes.Equal(repoData, localDataWithForcedUnixLineBreak)
	}

	return repoData, isDataIdentical, nil
}
