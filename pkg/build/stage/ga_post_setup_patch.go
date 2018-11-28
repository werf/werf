package stage

func NewGAPostSetupPatchStage() *GAPostSetupPatchStage {
	s := &GAPostSetupPatchStage{}
	s.GAStage = newGAStage()

	return s
}

type GAPostSetupPatchStage struct {
	*GAStage
}

func (s *GAPostSetupPatchStage) Name() StageName {
	return GAPostSetupPatch
}
