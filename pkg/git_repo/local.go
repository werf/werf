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
		fmt.Printf("Using commit `%s` of repo `%s`\n", commit, repo.String())
	}

	return commit, err
}
