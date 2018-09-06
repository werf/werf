package git_repo

import (
	"fmt"
	"io"
)

type Local struct {
	Base
	Path     string
	OrigPath string
}

func (repo *Local) HeadCommit() (string, error) {
	commit, err := repo.getHeadCommitForRepo(repo.Path)

	if err == nil {
		fmt.Printf("Using commit `%s` of repo `%s`\n", commit, repo.String())
	}

	return commit, err
}

func (repo *Local) CreatePatch(opts PatchOptions) (Patch, error) {
	return repo.createPatch(repo.Path, opts)
}

func (repo *Local) ArchiveType(opts ArchiveOptions) (ArchiveType, error) {
	return repo.archiveType(repo.Path, opts)
}

func (repo *Local) IsAnyEntries(opts ArchiveOptions) (bool, error) {
	return repo.isAnyEntries(repo.Path, opts)
}

func (repo *Local) CreateArchiveTar(output io.Writer, opts ArchiveOptions) error {
	return repo.createArchiveTar(repo.Path, output, opts)
}
