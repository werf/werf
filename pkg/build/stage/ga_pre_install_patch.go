package stage

func NewGAPreInstallPatchStage() *GAPreInstallPatchStage {
	s := &GAPreInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage()

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
