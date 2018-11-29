package stage

func newGARelatedStage() *GARelatedStage {
	s := &GARelatedStage{}
	s.GAStage = newGAStage()

	return s
}

type GARelatedStage struct {
	*GAStage
}

func (s *GARelatedStage) GetDependencies(_ Conveyor, _ Image) string {
	return ""
}

func (s *GARelatedStage) GetRelatedStageName() string {
	panic("method must be implemented!")
}
