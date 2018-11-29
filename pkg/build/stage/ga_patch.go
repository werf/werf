package stage

func newGAPatchStage() *GAPatchStage {
	s := &GAPatchStage{}
	return s
}

type GAPatchStage struct {
	*BaseStage
}

func (s *GAPatchStage) PrepareImage(prevImage, image Image) error {
	if err := s.BaseStage.PrepareImage(prevImage, image); err != nil {
		return err
	}

	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevImage, image); err != nil {
			return err
		}
	}

	return nil
}
