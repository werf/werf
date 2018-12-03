package stage

import "github.com/flant/dapp/pkg/image"

func newGAPatchStage(baseStageOptions *NewBaseStageOptions) *GAPatchStage {
	s := &GAPatchStage{}
	s.GAStage = newGAStage(baseStageOptions)
	return s
}

type GAPatchStage struct {
	*GAStage
}

func (s *GAPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if s.willLatestCommitBeBuiltOnPrevStage(prevBuiltImage) {
		return nil
	}

	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	return nil
}
