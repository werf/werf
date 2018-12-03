package stage

func newGAStage(baseStageOptions *NewBaseStageOptions) *GAStage {
	s := &GAStage{}
	s.BaseStage = newBaseStage(baseStageOptions)
	return s
}

type GAStage struct {
	*BaseStage
}
