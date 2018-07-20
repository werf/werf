package git_repo

import (
	"fmt"
)

type Local struct {
	Base
	Path     string
	OrigPath string
}

func (repo *Local) HeadCommit() (string, error) {
	commit, err := repo.getHeadCommitForRepo(repo.Path)

	if err == nil {
		fmt.Printf("Using commit `%s` of local repository\n", commit)
	}

	return commit, err
}
