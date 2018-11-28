package stage

func NewGALatestPatchStage() *GALatestPatchStage {
	s := &GALatestPatchStage{}
	s.GAStage = newGAStage()

	return s
}

type GALatestPatchStage struct {
	*GAStage
}

func (s *GALatestPatchStage) Name() StageName {
	return GALatestPatch
}
