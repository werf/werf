package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_inspector"
	imagePkg "github.com/werf/werf/pkg/image"
)

func newUserWithGitPatchStage(builder builder.Builder, name StageName, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *UserWithGitPatchStage {
	s := &UserWithGitPatchStage{}
	s.UserStage = newUserStage(builder, name, baseStageOptions)
	s.GitPatchStage = newGitPatchStage(name, gitPatchStageOptions, baseStageOptions)
	s.GitPatchStage.BaseStage = s.BaseStage

	return s
}

type UserWithGitPatchStage struct {
	*UserStage
	GitPatchStage *GitPatchStage
}

func (s *UserWithGitPatchStage) SelectSuitableStage(ctx context.Context, c Conveyor, stages []*imagePkg.StageDescription) (*imagePkg.StageDescription, error) {
	ancestorsImages, err := s.selectStagesAncestorsByGitMappings(ctx, c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsImages)
}

func (s *UserWithGitPatchStage) GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(ctx, c)
}

func (s *UserWithGitPatchStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.BaseStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	var withStagingPatch bool
	if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[imagePkg.WerfDevLabel] != "true" {
		empty, err := s.UserStage.isStageDependenciesStagingStatusChecksumEmpty(ctx, c, s.name)
		if err != nil {
			return err
		}

		withStagingPatch = !empty
	}

	if isPatchEmpty, err := s.GitPatchStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return err
	} else if !isPatchEmpty || withStagingPatch {
		if err := s.GitPatchStage.prepareImage(ctx, c, prevBuiltImage, image, withStagingPatch); err != nil {
			return err
		}
	}

	return nil
}
