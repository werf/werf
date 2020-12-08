package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_inspector"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

const patchSizeStep = 1024 * 1024

func NewGitCacheStage(gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitCacheStage {
	s := &GitCacheStage{}
	s.GitPatchStage = newGitPatchStage(GitCache, gitPatchStageOptions, baseStageOptions)
	return s
}

type GitCacheStage struct {
	*GitPatchStage
}

func (s *GitCacheStage) SelectSuitableStage(ctx context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	ancestorsImages, err := s.selectStagesAncestorsByGitMappings(ctx, c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsImages)
}

func (s *GitCacheStage) IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	if isEmptyBase, err := s.GitPatchStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return isEmptyBase, err
	} else if isEmptyBase {
		return true, err
	}

	patchSize, err := s.gitMappingsPatchSize(ctx, c, prevBuiltImage)
	if err != nil {
		return false, err
	}

	//if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[image.WerfDevLabel] != "true" {
	// TODO: s.gitMappingsStagingPatchSize
	//}

	isEmpty := patchSize < patchSizeStep

	return isEmpty, nil
}

func (s *GitCacheStage) GetDependencies(ctx context.Context, c Conveyor, _, prevBuiltImage container_runtime.ImageInterface) (string, error) {
	patchSize, err := s.gitMappingsPatchSize(ctx, c, prevBuiltImage)
	if err != nil {
		return "", err
	}

	//if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[image.WerfDevLabel] != "true" {
	// TODO: s.gitMappingsStagingPatchSize
	//}

	return util.Sha256Hash(fmt.Sprintf("%d", patchSize/patchSizeStep)), nil
}

func (s *GitCacheStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, i container_runtime.ImageInterface) error {
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

func (s *GitCacheStage) gitMappingsPatchSize(ctx context.Context, c Conveyor, prevBuiltImage container_runtime.ImageInterface) (int64, error) {
	var size int64
	for _, gitMapping := range s.gitMappings {
		commit, err := gitMapping.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
		if err != nil {
			return 0, fmt.Errorf("unable to get base commit for git mapping %s: %s", gitMapping.GitRepo().GetName(), err)
		}

		exist, err := gitMapping.GitRepo().IsCommitExists(ctx, commit)
		if err != nil {
			return 0, err
		}

		if exist {
			patchSize, err := gitMapping.PatchSize(ctx, c, commit)
			if err != nil {
				return 0, err
			}

			size += patchSize
		} else {
			return 0, nil
		}
	}

	return size, nil
}

func (s *GitCacheStage) GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(ctx, c)
}
