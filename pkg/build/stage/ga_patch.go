package stage

import (
	"fmt"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/image"
)

type NewGaPatchStageOptions struct {
	PatchesDir          string
	ContainerPatchesDir string
}

func newGAPatchStage(name StageName, gaPatchStageOptions *NewGaPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GAPatchStage {
	s := &GAPatchStage{
		PatchesDir:          gaPatchStageOptions.PatchesDir,
		ContainerPatchesDir: gaPatchStageOptions.ContainerPatchesDir,
	}
	s.GAStage = newGAStage(name, baseStageOptions)
	return s
}

type GAPatchStage struct {
	*GAStage

	PatchesDir          string
	ContainerPatchesDir string
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
	if err := s.GAStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.prepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}

func (s *GAPatchStage) prepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	for _, ga := range s.gitArtifacts {
		if err := ga.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	gitArtifactContainerName, err := dappdeps.GitArtifactContainer()
	if err != nil {
		return err
	}

	image.Container().RunOptions().AddVolumeFrom(gitArtifactContainerName)
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.PatchesDir, s.ContainerPatchesDir))

	return nil
}
