package git_repo

import (
	"fmt"
	"path/filepath"
)

type Local struct {
	Base
	Path   string
	GitDir string
}

func (repo *Local) HeadCommit() (string, error) {
	commit, err := repo.getHeadCommitForRepo(repo.Path)

	if err == nil {
		fmt.Printf("Using commit `%s` of repo `%s`\n", commit, repo.String())
	}

	return commit, err
}

func (repo *Local) CreatePatch(opts PatchOptions) (Patch, error) {
	return repo.createPatch(repo.Path, repo.GitDir, repo.getWorkTreeDir(), opts)
}

func (repo *Local) CreateArchive(opts ArchiveOptions) (Archive, error) {
	return repo.createArchive(repo.Path, repo.GitDir, repo.getWorkTreeDir(), opts)
}

func (repo *Local) IsCommitExists(commit string) (bool, error) {
	return repo.isCommitExists(repo.Path, commit)
}

func (repo *Local) TagsList() ([]string, error) {
	return repo.tagsList(repo.Path)
}

func (repo *Local) RemoteBranchesList() ([]string, error) {
	return repo.remoteBranchesList(repo.Path)
}

func (repo *Local) getWorkTreeDir() string {
	pathParts := make([]string, 0)

	path := filepath.Clean(repo.Path)
	for i := 0; i < 3; i++ {
		var lastPart string
		path, lastPart = filepath.Split(filepath.Clean(path))
		pathParts = append([]string{lastPart}, pathParts...)
		if path == "/" {
			break
		}
	}

	pathParts = append([]string{"local"}, pathParts...)
	pathParts = append([]string{GetBaseWorkTreeDir()}, pathParts...)

	return filepath.Join(pathParts...)
}
