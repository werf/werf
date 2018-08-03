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

func (repo *Local) Diff(basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (string, error) {
	return repo.diff(repo.Path, basePath, fromCommit, toCommit, includePaths, excludePaths)
}

func (repo *Local) IsAnyChanges(basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (bool, error) {
	return repo.isAnyChanges(repo.Path, basePath, fromCommit, toCommit, includePaths, excludePaths)
}
