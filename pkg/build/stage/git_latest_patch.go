package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_inspector"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

func NewGitLatestPatchStage(gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitLatestPatchStage {
	s := &GitLatestPatchStage{}
	s.GitPatchStage = newGitPatchStage(GitLatestPatch, gitPatchStageOptions, baseStageOptions)
	return s
}

type GitLatestPatchStage struct {
	*GitPatchStage
}

func (s *GitLatestPatchStage) IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	if empty, err := s.GitPatchStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	isEmpty := true
	for _, gm := range s.gitMappings {
		commit, err := gm.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
		if err != nil {
			return false, fmt.Errorf("unable to get base commit for git mapping %s: %s", gm.GitRepo().GetName(), err)
		}

		if exist, err := gm.GitRepo().IsCommitExists(ctx, commit); err != nil {
			return false, fmt.Errorf("unable to check existence of commit %q in the repo %s: %s", commit, gm.GitRepo().GetName(), err)
		} else if !exist {
			return false, fmt.Errorf("commit %q is not exist in the repo %s", commit, gm.GitRepo().GetName())
		}

		if empty, err := gm.IsPatchEmpty(ctx, c, prevBuiltImage); err != nil {
			return false, err
		} else if !empty {
			isEmpty = false
			break
		}

		if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[image.WerfDevLabel] != "true" {
			empty, err := gm.IsStagingStatusResultEmpty(ctx)
			if err != nil {
				return false, err
			}

			if !empty {
				isEmpty = false
				break
			}
		}
	}

	return isEmpty, nil
}

func (s *GitLatestPatchStage) GetDependencies(ctx context.Context, c Conveyor, _, prevBuiltImage container_runtime.ImageInterface) (string, error) {
	var args []string

	for _, gm := range s.gitMappings {
		patchContent, err := gm.GetPatchContent(ctx, c, prevBuiltImage)
		if err != nil {
			return "", fmt.Errorf("error getting patch between previous built image %s and current commit for git mapping %s: %s", prevBuiltImage.Name(), gm.Name, err)
		}

		args = append(args, patchContent)
	}

	if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[image.WerfDevLabel] != "true" {
		devModeChecksum, err := s.gitMappingStagingStatusChecksum(ctx)
		if err != nil {
			return "", err
		}

		args = append(args, devModeChecksum)
	}

	return util.Sha256Hash(args...), nil
}

func (s *GitLatestPatchStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, i container_runtime.ImageInterface) error {
	var withStagingPatch bool
	if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[image.WerfDevLabel] != "true" {
		empty, err := s.isGitMappingStagingStatusChecksumEmpty(ctx)
		if err != nil {
			return err
		}

		withStagingPatch = !empty
	}

	return s.GitPatchStage.PrepareImage(ctx, c, prevBuiltImage, i, withStagingPatch)
}

func (s *GitLatestPatchStage) SelectSuitableStage(ctx context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	ancestorsImages, err := s.selectStagesAncestorsByGitMappings(ctx, c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsImages)
}

func (s *GitLatestPatchStage) GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(ctx, c)
}
