package stage

import (
	"github.com/flant/dapp/pkg/image"
)

func newGAPatchStage(baseStageOptions *NewBaseStageOptions) *GAPatchStage {
	s := &GAPatchStage{}
	s.GAStage = newGAStage(baseStageOptions)
	return s
}

type GAPatchStage struct {
	*GAStage
}

func (s *GAPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if s.willLatestCommitBeBuiltOnPrevStage(prevBuiltImage) {
		return nil
	}

	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	return nil
}

func (s *GAPatchStage) willLatestCommitBeBuiltOnPrevStage(prevBuiltImage image.Image) bool {
	if prevBuiltImage == nil {
		return true
	}

	for _, ga := range s.gitArtifacts {
		_, exist := prevBuiltImage.Labels()[ga.GetParamshash()]
		if !exist {
			return true
		}
	}

	return true
}
