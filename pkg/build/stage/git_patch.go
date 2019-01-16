package stage

import (
	"fmt"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/image"
)

type NewGitPatchStageOptions struct {
	PatchesDir          string
	ContainerPatchesDir string
}

func newGitPatchStage(name StageName, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitPatchStage {
	s := &GitPatchStage{
		PatchesDir:          gitPatchStageOptions.PatchesDir,
		ContainerPatchesDir: gitPatchStageOptions.ContainerPatchesDir,
	}
	s.GitStage = newGitStage(name, baseStageOptions)
	return s
}

type GitPatchStage struct {
	*GitStage

	PatchesDir          string
	ContainerPatchesDir string
}

func (s *GitPatchStage) IsEmpty(c Conveyor, prevBuiltImage image.Image) (bool, error) {
	if empty, err := s.willGitLatestCommitBeBuiltOnPrevGitStage(c); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	if empty, err := s.GitStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	if empty, err := s.hasPrevBuiltStageHadActualGitPaths(prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return false, nil
}

func (s *GitPatchStage) willGitLatestCommitBeBuiltOnPrevGitStage(c Conveyor) (bool, error) {
	stageName := c.GetBuildingGitStage(s.dimgName)
	if stageName != "" && stageName != s.Name() {
		return true, nil
	}

	return false, nil
}

func (s *GitPatchStage) hasPrevBuiltStageHadActualGitPaths(prevBuiltImage image.Image) (bool, error) {
	for _, gitPath := range s.gitPaths {
		commit := gitPath.GetGitCommitFromImageLabels(prevBuiltImage)
		latestCommit, err := gitPath.LatestCommit()
		if err != nil {
			return false, err
		}

		if commit != latestCommit {
			return false, nil
		}
	}

	return true, nil
}

func (s *GitPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.GitStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.prepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}

func (s *GitPatchStage) prepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	for _, gitPath := range s.gitPaths {
		if err := gitPath.ApplyPatchCommand(prevBuiltImage, image); err != nil {
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
