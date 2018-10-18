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

func (repo *Local) FindCommitIdByMessage(regex string) (string, error) {
	head, err := repo.HeadCommit()
	if err != nil {
		return "", fmt.Errorf("error getting head commit: %s", err)
	}
	return repo.findCommitIdByMessage(repo.Path, regex, head)
}

func (repo *Local) IsEmpty() (bool, error) {
	return repo.isEmpty(repo.Path)
}

func (repo *Local) RemoteOriginUrl() (string, error) {
	return repo.remoteOriginUrl(repo.Path)
}

func (repo *Local) HeadCommit() (string, error) {
	ref, err := repo.getReferenceForRepo(repo.Path)
	if err != nil {
		return "", fmt.Errorf("cannot get repo `%s` head ref: %s", repo.Path, err)
	}

	commit := fmt.Sprintf("%s", ref.Hash())

	fmt.Printf("Using HEAD commit `%s` of repo `%s`\n", commit, repo.String())

	return commit, nil
}

func (repo *Local) HeadBranchName() (string, error) {
	return repo.getHeadBranchName(repo.Path)
}

func (repo *Local) CreatePatch(opts PatchOptions) (Patch, error) {
	return repo.createPatch(repo.Path, repo.GitDir, repo.getWorkTreeDir(), opts)
}

func (repo *Local) CreateArchive(opts ArchiveOptions) (Archive, error) {
	return repo.createArchive(repo.Path, repo.GitDir, repo.getWorkTreeDir(), opts)
}

func (repo *Local) Checksum(opts ChecksumOptions) (Checksum, error) {
	return repo.checksum(repo.Path, repo.GitDir, repo.getWorkTreeDir(), opts)
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
