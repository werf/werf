package stage

import "github.com/flant/dapp/pkg/image"

func newGARelatedStage(baseStageOptions *NewBaseStageOptions) *GARelatedStage {
	s := &GARelatedStage{}
	s.GAPatchStage = newGAPatchStage(baseStageOptions)
	return s
}

type GARelatedStage struct {
	*GAPatchStage
}

func (s *GARelatedStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	return "", nil
}

func (s *GARelatedStage) GetRelatedStageName() StageName {
	panic("method must be implemented!")
}
