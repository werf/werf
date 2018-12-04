package stage

func NewGAPreInstallPatchStage(baseStageOptions *NewBaseStageOptions) *GAPreInstallPatchStage {
	s := &GAPreInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage(baseStageOptions)
	return s
}

type GAPreInstallPatchStage struct {
	*GARelatedStage
}

func (s *GAPreInstallPatchStage) Name() StageName {
	return GAPreInstallPatch
}

func (s *GAPreInstallPatchStage) GetRelatedStageName() StageName {
	return Install
}
