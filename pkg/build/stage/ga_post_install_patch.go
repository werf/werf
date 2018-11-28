package stage

func NewGAPostInstallPatchStage() *GAPostInstallPatchStage {
	s := &GAPostInstallPatchStage{}
	s.GARelatedStage = newGARelatedStage()

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
