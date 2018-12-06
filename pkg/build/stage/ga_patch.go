package stage

import (
	"github.com/flant/dapp/pkg/image"
)

func newGAPatchStage(name StageName, baseStageOptions *NewBaseStageOptions) *GAPatchStage {
	s := &GAPatchStage{}
	s.GAStage = newGAStage(name, baseStageOptions)
	return s
}

type GAPatchStage struct {
	*GAStage
}

func (s *GAPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if s.willLatestCommitBeBuiltOnGAArchiveStage(prevBuiltImage) {
		return nil
	}

	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	return nil
}

func (s *GAPatchStage) willLatestCommitBeBuiltOnGAArchiveStage(prevBuiltImage image.Image) bool {
	if prevBuiltImage == nil {
		return true
	}

	for _, ga := range s.gitArtifacts {
		if ga.GetGACommitFromImageLabels(prevBuiltImage) == "" {
			return true
		}
	}

	return false
}
