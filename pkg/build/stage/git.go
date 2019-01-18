package stage

import (
	"github.com/flant/werf/pkg/image"
)

func newGitStage(name StageName, baseStageOptions *NewBaseStageOptions) *GitStage {
	s := &GitStage{}
	s.BaseStage = newBaseStage(name, baseStageOptions)
	return s
}

type GitStage struct {
	*BaseStage
}

func (s *GitStage) IsEmpty(_ Conveyor, prevBuiltImage image.Image) (bool, error) {
	return len(s.gitPaths) == 0, nil
}

func (s *GitStage) ShouldBeReset(builtImage image.Image) (bool, error) {
	for _, gitPath := range s.gitPaths {
		commit := gitPath.GetGitCommitFromImageLabels(builtImage)
		if exist, err := gitPath.GitRepo().IsCommitExists(commit); err != nil {
			return false, err
		} else if !exist {
			return true, nil
		}
	}

	return false, nil
}

func (s *GitStage) AfterImageSyncDockerStateHook(c Conveyor) error {
	if !s.image.IsExists() {
		stageName := c.GetBuildingGitStage(s.dimgName)
		if stageName == "" {
			c.SetBuildingGitStage(s.dimgName, s.Name())
		}
	}

	return nil
}

func (s *GitStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}
