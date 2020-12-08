package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_inspector"
	imagePkg "github.com/werf/werf/pkg/image"
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

func (s *GitPatchStage) IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	if empty, err := s.GitStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return s.hasPrevBuiltStageHadActualGitMappings(ctx, c, prevBuiltImage)
}

func (s *GitPatchStage) hasPrevBuiltStageHadActualGitMappings(ctx context.Context, c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
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

		if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[imagePkg.WerfDevLabel] != "true" {
			empty, err := gitMapping.IsStagingStatusResultEmpty(ctx)
			if err != nil {
				return false, err
			}

			if !empty {
				return false, nil
			}
		}
	}

	return true, nil
}

func (s *GitPatchStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface, withIndexPatch bool) error {
	if err := s.GitStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.prepareImage(ctx, c, prevBuiltImage, image, withIndexPatch); err != nil {
		return err
	}

	return nil
}

func (s *GitPatchStage) prepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface, withStagingPatch bool) error {
	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyPatchCommand(ctx, c, prevBuiltImage, image); err != nil {
			return err
		}

		if withStagingPatch {
			// TODO gitMapping.ApplyStagingPatchCommand
		}
	}

	if withStagingPatch {
		image.Container().RunOptions().AddLabel(map[string]string{imagePkg.WerfDevLabel: "true"})
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", git_repo.CommonGitDataManager.PatchesCacheDir, s.ContainerPatchesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", git_repo.CommonGitDataManager.ArchivesCacheDir, s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
