package stage

func newGAStage(name StageName, baseStageOptions *NewBaseStageOptions) *GAStage {
	s := &GAStage{}
	s.BaseStage = newBaseStage(name, baseStageOptions)
	return s
}

type GAStage struct {
	*BaseStage
}
