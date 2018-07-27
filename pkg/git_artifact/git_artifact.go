package git_artifact

import (
	"fmt"
	"github.com/flant/dapp/pkg/git_repo"
)

type GitArtifact struct {
	LocalGitRepo  *git_repo.Local
	RemoteGitRepo *git_repo.Remote

	Name               string
	As                 string
	Branch             string
	Tag                string
	Commit             string
	Cwd                string
	Owner              string
	Group              string
	IncludePaths       []string
	ExcludePaths       []string
	StagesDependencies map[string][]string
	Paramshash         string
}

func (ga *GitArtifact) GitRepo() (git_repo.GitRepo, error) {
	if ga.LocalGitRepo != nil {
		return ga.LocalGitRepo, nil
	} else if ga.RemoteGitRepo != nil {
		return ga.RemoteGitRepo, nil
	}
	return nil, fmt.Errorf("GitRepo not initialized")
}

func (ga *GitArtifact) IsLocal() bool {
	if ga.LocalGitRepo != nil {
		return true
	} else {
		return false
	}
}

func (ga *GitArtifact) LatestCommit() (string, error) {
	gitRepo, err := ga.GitRepo()
	if err != nil {
		return "", err
	}

	if ga.Commit != "" {
		fmt.Printf("Using specified commit `%s` of repository `%s`\n", ga.Commit, gitRepo.String())
		return ga.Commit, nil
	}

	if ga.Tag != "" {
		return gitRepo.LatestTagCommit(ga.Tag)
	}

	if ga.Branch != "" {
		return gitRepo.LatestBranchCommit(ga.Branch)
	}

	if ga.IsLocal() {
		return gitRepo.HeadCommit()
	} else {
		branchName, err := gitRepo.HeadBranchName()
		if err != nil {
			return "", err
		}
		return gitRepo.LatestBranchCommit(branchName)
	}
}
