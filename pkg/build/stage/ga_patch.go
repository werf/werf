package stage

func newGAPatchStage() *GAPatchStage {
	s := &GAPatchStage{}
	s.GAStage = newGAStage()
	return s
}

type GAPatchStage struct {
	*GAStage
}

func (s *GAPatchStage) PrepareImage(prevBuiltImage, image Image) error {
	if s.willLatestCommitBeBuiltOnPrevStage(prevBuiltImage) {
		return nil
	}

	if err := s.BaseStage.PrepareImage(prevBuiltImage, image); err != nil {
		return err
	}

	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	return nil
}
