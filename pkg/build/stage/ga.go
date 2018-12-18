package stage

import (
	"github.com/flant/dapp/pkg/image"
)

func newGAStage(name StageName, baseStageOptions *NewBaseStageOptions) *GAStage {
	s := &GAStage{}
	s.BaseStage = newBaseStage(name, baseStageOptions)
	return s
}

type GAStage struct {
	*BaseStage
}

func (s *GAStage) IsEmpty(_ Conveyor, prevBuiltImage image.Image) (bool, error) {
	return len(s.gitArtifacts) == 0, nil
}

func (s *GAStage) ShouldBeReset(builtImage image.Image) (bool, error) {
	for _, ga := range s.gitArtifacts {
		commit := ga.GetGACommitFromImageLabels(builtImage)
		if exist, err := ga.GitRepo().IsCommitExists(commit); err != nil {
			return false, err
		} else if !exist {
			return true, nil
		}
	}

	return false, nil
}

func (s *GAStage) AfterImageSyncDockerStateHook(c Conveyor) error {
	if !s.image.IsExists() {
		stageName := c.GetBuildingGAStage(s.dimgName)
		if stageName == "" {
			c.SetBuildingGAStage(s.dimgName, s.Name())
		}
	}

	return nil
}

func (s *GAStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}
