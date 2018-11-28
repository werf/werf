package stage

func newGAStage() *GAStage {
	s := &GAStage{}
	return s
}

type GAStage struct {
	*BaseStage
}
