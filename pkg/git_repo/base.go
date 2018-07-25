package git_repo

import (
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Base struct {
	Name string
}

func (repo *Base) getHeadCommitForRepo(repoPath string) (string, error) {
	ref, err := repo.getReferenceForRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	return fmt.Sprintf("%s", ref.Hash()), nil
}

func (repo *Base) getHeadBranchNameForRepo(repoPath string) (string, error) {
	ref, err := repo.getReferenceForRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	if ref.Name().IsBranch() {
		branchRef := ref.Name()
		return strings.Split(string(branchRef), "refs/heads/")[1], nil
	} else {
		return "", fmt.Errorf("cannot get branch name: HEAD refers to a specific revision that is not associated with a branch name")
	}
}

func (repo *Base) getReferenceForRepo(repoPath string) (*plumbing.Reference, error) {
	var err error

	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	return repository.Head()
}

func (repo *Base) String() string {
	return repo.Name
}

func (repo *Base) HeadCommit() (string, error) {
	panic("not implemented")
}

func (repo *Base) HeadBranchName() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (repo *Base) LatestBranchCommit(branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) LatestTagCommit(branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) Diff(path string, fromCommit, toCommit string, includePaths, excludePaths []string) (string, error) {
	panic("not implemented")
}

func (repo *Base) IsAnyChanges(basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (bool, error) {
	panic("not implemented")
}

func (repo *Base) makePatch(repoPath string, basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (*RelativeFilteredPatch, error) {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	fromHash := plumbing.NewHash(fromCommit)
	fromCommitObj, err := repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to  `%s`: %s", fromCommit, err)
	}

	toHash := plumbing.NewHash(toCommit)
	toCommitObj, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("cannot create tree for commit `%s`: %s", toCommit, err)
	}

	originPatch, err := fromCommitObj.Patch(toCommitObj)
	if err != nil {
		return nil, fmt.Errorf("cannot create patch between `%s` and `%s`: %s", fromCommit, toCommit, err)
	}

	return &RelativeFilteredPatch{
		OriginPatch:  originPatch,
		BasePath:     basePath,
		IncludePaths: includePaths,
		ExcludePaths: excludePaths,
	}, nil
}

func (repo *Base) diff(repoPath string, basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (string, error) {
	patch, err := repo.makePatch(repoPath, basePath, fromCommit, toCommit, includePaths, excludePaths)

	if err != nil {
		return "", err
	}

	return patch.String(), nil
}

func (repo *Base) isAnyChanges(repoPath string, basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (bool, error) {
	patch, err := repo.makePatch(repoPath, basePath, fromCommit, toCommit, includePaths, excludePaths)

	if err != nil {
		return false, err
	}

	return (len(patch.FilePatches()) != 0), nil
}
