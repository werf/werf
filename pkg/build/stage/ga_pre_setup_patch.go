package stage

func NewGAPreSetupPatchStage() *GAPreInstallPatchStage {
	s := &GAPreInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage()

	return s
}

type GAPreSetupPatchStage struct {
	*GARelatedStage
}

func (s *GAPreSetupPatchStage) Name() StageName {
	return GAPreSetupPatch
}

func (s *GAPreSetupPatchStage) GetRelatedStageName() StageName {
	return Setup
}
