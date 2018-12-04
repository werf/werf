package stage

func NewGAPostInstallPatchStage(baseStageOptions *NewBaseStageOptions) *GAPostInstallPatchStage {
	s := &GAPostInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage(baseStageOptions)
	return s
}

type GAPostInstallPatchStage struct {
	*GARelatedStage
}

func (s *GAPostInstallPatchStage) Name() StageName {
	return GAPostInstallPatch
}

func (s *GAPostInstallPatchStage) GetRelatedStageName() StageName {
	return BeforeSetup
}
