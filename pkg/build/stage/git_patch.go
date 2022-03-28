package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/git_repo"
)

type NewGitPatchStageOptions struct {
	ScriptsDir           string
	ContainerPatchesDir  string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func newGitPatchStage(name StageName, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitPatchStage {
	s := &GitPatchStage{
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

	ScriptsDir           string
	ContainerPatchesDir  string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func (s *GitPatchStage) IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (bool, error) {
	if empty, err := s.GitStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return s.hasPrevBuiltStageHadActualGitMappings(ctx, c, prevBuiltImage)
}

func (s *GitPatchStage) hasPrevBuiltStageHadActualGitMappings(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (bool, error) {
	for _, gitMapping := range s.gitMappings {
		commit, err := gitMapping.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
		if err != nil {
			return false, fmt.Errorf("unable to get base commit for git mapping %s: %s", gitMapping.GitRepo().GetName(), err)
		}

		latestCommitInfo, err := gitMapping.GetLatestCommitInfo(ctx, c)
		if err != nil {
			return false, fmt.Errorf("unable to get latest commit for git mapping %s: %s", gitMapping.GitRepo().GetName(), err)
		}

		if commit != latestCommitInfo.Commit {
			return false, nil
		}
	}

	return true, nil
}

func (s *GitPatchStage) PrepareImage(ctx context.Context, c Conveyor, cr container_runtime.ContainerRuntime, prevBuiltImage, stageImage *StageImage) error {
	if err := s.GitStage.PrepareImage(ctx, c, cr, prevBuiltImage, stageImage); err != nil {
		return err
	}

	if err := s.prepareImage(ctx, c, cr, prevBuiltImage, stageImage); err != nil {
		return err
	}

	return nil
}

func (s *GitPatchStage) prepareImage(ctx context.Context, c Conveyor, cr container_runtime.ContainerRuntime, prevBuiltImage, stageImage *StageImage) error {
	if cr.HasContainerRootMountSupport() {
		// TODO(stapel-to-buildah)
		panic("not implemented")
	} else {
		for _, gitMapping := range s.gitMappings {
			if err := gitMapping.ApplyPatchCommand(ctx, c, prevBuiltImage, stageImage); err != nil {
				return err
			}
		}

		stageImage.StageBuilderAccessor.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", git_repo.CommonGitDataManager.GetPatchesCacheDir(), s.ContainerPatchesDir))
		stageImage.StageBuilderAccessor.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", git_repo.CommonGitDataManager.GetArchivesCacheDir(), s.ContainerArchivesDir))
		stageImage.StageBuilderAccessor.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

		return nil
	}
}
