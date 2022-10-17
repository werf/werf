package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

const patchSizeStep = 1024 * 1024

func NewGitCacheStage(gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *GitCacheStage {
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
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %w", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsImages)
}

func (s *GitCacheStage) IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (bool, error) {
	if isEmptyBase, err := s.GitPatchStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return isEmptyBase, err
	} else if isEmptyBase {
		return true, err
	}

	patchSize, err := s.gitMappingsPatchSize(ctx, c, prevBuiltImage)
	if err != nil {
		return false, err
	}

	isEmpty := patchSize < patchSizeStep

	return isEmpty, nil
}

func (s *GitCacheStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	patchSize, err := s.gitMappingsPatchSize(ctx, c, prevBuiltImage)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(fmt.Sprintf("%d", patchSize/patchSizeStep)), nil
}

func (s *GitCacheStage) gitMappingsPatchSize(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (int64, error) {
	var size int64
	for _, gitMapping := range s.gitMappings {
		commit, err := gitMapping.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
		if err != nil {
			return 0, fmt.Errorf("unable to get base commit for git mapping %s: %w", gitMapping.GitRepo().GetName(), err)
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
