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

func (s *GAPatchStage) IsEmpty(c Conveyor, prevBuiltImage image.Image) (bool, error) {
	if empty, err := s.willGABeBuiltOnPrevGAStage(c); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	if empty, err := s.GAStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	if empty, err := s.hasPrevBuiltStageHadActualGA(prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return false, nil
}

func (s *GAPatchStage) willGABeBuiltOnPrevGAStage(c Conveyor) (bool, error) {
	stageName := c.GetBuildingGAStage(s.dimgName)
	if stageName != "" && stageName != s.Name() {
		return true, nil
	}

	return false, nil
}

func (s *GAPatchStage) hasPrevBuiltStageHadActualGA(prevBuiltImage image.Image) (bool, error) {
	for _, ga := range s.gitArtifacts {
		commit := ga.GetGACommitFromImageLabels(prevBuiltImage)
		latestCommit, err := ga.LatestCommit()
		if err != nil {
			return false, err
		}

		if commit != latestCommit {
			return false, nil
		}
	}

	return true, nil
}

func (s *GAPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	return nil
}
