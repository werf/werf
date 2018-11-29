package stage

func newGARelatedStage() *GARelatedStage {
	s := &GARelatedStage{}
	s.GAStage = newGAStage()

	return s
}

type GARelatedStage struct {
	*GAStage
}

func (s *GARelatedStage) GetDependencies(_ Conveyor, _ Image) (string, error) {
	return "", nil
}

func (s *GARelatedStage) GetRelatedStageName() StageName {
	panic("method must be implemented!")
}
