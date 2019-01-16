package stage

import (
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func NewGitLatestPatchStage(gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitLatestPatchStage {
	s := &GitLatestPatchStage{}
	s.GitPatchStage = newGitPatchStage(GitLatestPatch, gitPatchStageOptions, baseStageOptions)
	return s
}

type GitLatestPatchStage struct {
	*GitPatchStage
}

func (s *GitLatestPatchStage) IsEmpty(c Conveyor, prevBuiltImage image.Image) (bool, error) {
	if empty, err := s.GitPatchStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	isEmpty := true
	for _, gitPath := range s.gitPaths {
		commit := gitPath.GetGitCommitFromImageLabels(prevBuiltImage)
		if exist, err := gitPath.GitRepo().IsCommitExists(commit); err != nil {
			return false, err
		} else if !exist {
			return true, nil
		}

		if empty, err := gitPath.IsPatchEmpty(prevBuiltImage); err != nil {
			return false, err
		} else if !empty {
			isEmpty = false
			break
		}
	}

	return isEmpty, nil
}

func (s *GitLatestPatchStage) GetDependencies(_ Conveyor, prevImage image.Image) (string, error) {
	var args []string

	for _, gitPath := range s.gitPaths {
		commit, err := gitPath.LatestCommit()
		if err != nil {
			return "", err
		}

		args = append(args, commit)
	}

	return util.Sha256Hash(args...), nil
}
