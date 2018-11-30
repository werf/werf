package stage

func newGAStage() *GAStage {
	s := &GAStage{}
	s.BaseStage = newBaseStage()
	return s
}

type GAStage struct {
	*BaseStage
}

func (s *GAStage) willLatestCommitBeBuiltOnPrevStage(prevBuiltImage Image) bool {
	if prevBuiltImage == nil {
		return true
	}

	for _, ga := range s.gitArtifacts {
		_, exist := prevBuiltImage.Labels()[ga.GetParamshash()]
		if !exist {
			return false
		}
	}

	return true
}
