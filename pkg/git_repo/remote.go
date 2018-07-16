package git_repo

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"os"
)

type Remote struct {
	Base
	Url       string
	ClonePath string // TODO: move CacheVersion & path construction here
}

func (repo *Remote) clone() error {
	var err error

	_, err = os.Stat(repo.ClonePath)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("cannot clone git repo: %s", err)
	}

	fmt.Printf("Clone remote git repo `%s`\n", repo.Url)

	_, err = git.PlainClone(repo.ClonePath, true, &git.CloneOptions{
		URL:               repo.Url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return fmt.Errorf("cannot clone git repo: %s", err)
	}

	return nil
}

func (repo *Remote) HeadCommit() (string, error) {
	var err error

	err = repo.clone()
	if err != nil {
		return "", err
	}

	commit, err := repo.getHeadCommitForRepo(repo.ClonePath)

	if err == nil {
		fmt.Printf("Using commit `%s` of repository `%s`\n", commit, repo.Url)
	}

	return commit, err
}

func (repo *Remote) findReference(rawRepo *git.Repository, reference string) (string, error) {
	refs, err := rawRepo.References()
	if err != nil {
		return "", err
	}

	var res string

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().String() == reference {
			res = fmt.Sprintf("%s", ref.Hash())
			return storer.ErrStop
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return res, nil
}

func (repo *Remote) LatestBranchCommit(branch string) (string, error) {
	var err error

	err = repo.clone()
	if err != nil {
		return "", err
	}

	rawRepo, err := git.PlainOpen(repo.ClonePath)
	if err != nil {
		return "", fmt.Errorf("cannot open repo: %s", err)
	}

	res, err := repo.findReference(rawRepo, fmt.Sprintf("refs/remotes/origin/%s", branch))
	if err != nil {
		return "", err
	}
	if res == "" {
		return "", fmt.Errorf("unknown branch `%s` of repository `%s`", branch, repo.Url)
	}

	fmt.Printf("Using commit `%s` of repository `%s` branch `%s`\n", res, repo.Url, branch)

	return res, nil
}

func (repo *Remote) LatestTagCommit(tag string) (string, error) {
	var err error

	err = repo.clone()
	if err != nil {
		return "", err
	}

	rawRepo, err := git.PlainOpen(repo.ClonePath)
	if err != nil {
		return "", fmt.Errorf("cannot open repo: %s", err)
	}

	res, err := repo.findReference(rawRepo, fmt.Sprintf("refs/tags/%s", tag))
	if err != nil {
		return "", err
	}
	if res == "" {
		return "", fmt.Errorf("unknown tag `%s` of repository `%s`", tag, repo.Url)
	}

	fmt.Printf("Using commit `%s` of repository `%s` tag `%s`\n", res, repo.Url, tag)

	return res, nil
}
