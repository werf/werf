package stage

import (
	"fmt"

	"github.com/flant/werf/pkg/image"
)

type NewGitPatchStageOptions struct {
	PatchesDir           string
	ArchivesDir          string
	ScriptsDir           string
	ContainerPatchesDir  string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func newGitPatchStage(name StageName, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitPatchStage {
	s := &GitPatchStage{
		PatchesDir:           gitPatchStageOptions.PatchesDir,
		ArchivesDir:          gitPatchStageOptions.ArchivesDir,
		ScriptsDir:           gitPatchStageOptions.ScriptsDir,
		ContainerPatchesDir:  gitPatchStageOptions.ContainerPatchesDir,
		ContainerArchivesDir: gitPatchStageOptions.ContainerArchivesDir,
		ContainerScriptsDir:  gitPatchStageOptions.ContainerScriptsDir,
	}
	s.GitStage = newGitStage(name, baseStageOptions)
	return s
}

type GitPatchStage struct {
	*GitStage

	PatchesDir           string
	ArchivesDir          string
	ScriptsDir           string
	ContainerPatchesDir  string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func (s *GitPatchStage) IsEmpty(c Conveyor, prevBuiltImage image.ImageInterface) (bool, error) {
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

	if empty, err := s.hasPrevBuiltStageHadActualGitMappings(prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return false, nil
}

func (s *GitPatchStage) willGitLatestCommitBeBuiltOnPrevGitStage(c Conveyor) (bool, error) {
	stageName := c.GetBuildingGitStage(s.imageName)
	if stageName != "" && stageName != s.Name() {
		return true, nil
	}

	return false, nil
}

func (s *GitPatchStage) hasPrevBuiltStageHadActualGitMappings(prevBuiltImage image.ImageInterface) (bool, error) {
	for _, gitMapping := range s.gitMappings {
		commit := gitMapping.GetGitCommitFromImageLabels(prevBuiltImage)
		latestCommit, err := gitMapping.LatestCommit()
		if err != nil {
			return false, err
		}

		if commit != latestCommit {
			return false, nil
		}
	}

	return true, nil
}

func (s *GitPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	if err := s.GitStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.prepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}

func (s *GitPatchStage) prepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyPatchCommand(prevBuiltImage, image); err != nil {
			return err
		}
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.PatchesDir, s.ContainerPatchesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ArchivesDir, s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
