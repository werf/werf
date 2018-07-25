package git_repo

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
)

type Base struct {
	Name string
}

func (repo *Base) getHeadCommitForRepo(repoPath string) (string, error) {
	var err error

	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot open repo: %s", err)
	}

	ref, err := repository.Head()
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	return fmt.Sprintf("%s", ref.Hash()), nil
}

func (repo *Base) String() string {
	return repo.Name
}

func (repo *Base) HeadCommit() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (repo *Base) LatestBranchCommit(branch string) (string, error) {
	return "", fmt.Errorf("not implemeneted")
}

func (repo *Base) LatestTagCommit(branch string) (string, error) {
	return "", fmt.Errorf("not implemeneted")
}
