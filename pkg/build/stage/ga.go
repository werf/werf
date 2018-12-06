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

func (s *GAStage) AfterImageSyncDockerStateHook(c Conveyor) error {
	if !s.image.IsExists() {
		stageName := c.GetBuildingGAStage(s.dimgName)
		if stageName == "" {
			c.SetBuildingGAStage(s.dimgName, s.Name())
		}
	}

	return nil
}
