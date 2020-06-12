package stage

import (
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
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

func (s *GitPatchStage) IsEmpty(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	if empty, err := s.GitStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return s.hasPrevBuiltStageHadActualGitMappings(c, prevBuiltImage)
}

func (s *GitPatchStage) hasPrevBuiltStageHadActualGitMappings(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	for _, gitMapping := range s.gitMappings {
		commit, err := gitMapping.GetBaseCommitForPrevBuiltImage(c, prevBuiltImage)
		if err != nil {
			return false, fmt.Errorf("unable to get base commit for git mapping %s: %s", gitMapping.GitRepo().GetName(), err)
		}

		latestCommitInfo, err := gitMapping.GetLatestCommitInfo(c)
		if err != nil {
			return false, fmt.Errorf("unable to get latest commit for git mapping %s: %s", gitMapping.GitRepo().GetName(), err)
		}

		if commit != latestCommitInfo.Commit {
			return false, nil
		}
	}

	return true, nil
}

func (s *GitPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.GitStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.prepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}

func (s *GitPatchStage) prepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyPatchCommand(c, prevBuiltImage, image); err != nil {
			return err
		}
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.PatchesDir, s.ContainerPatchesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ArchivesDir, s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
