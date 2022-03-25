package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/image"
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

func (s *UserWithGitPatchStage) SelectSuitableStage(ctx context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	ancestorsImages, err := s.selectStagesAncestorsByGitMappings(ctx, c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsImages)
}

func (s *UserWithGitPatchStage) GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(ctx, c)
}

func (s *UserWithGitPatchStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, stageImage *StageImage) error {
	if err := s.BaseStage.PrepareImage(ctx, c, prevBuiltImage, stageImage); err != nil {
		return err
	}

	if isPatchEmpty, err := s.GitPatchStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return err
	} else if !isPatchEmpty {
		if err := s.GitPatchStage.prepareImage(ctx, c, prevBuiltImage, stageImage); err != nil {
			return err
		}
	}

	return nil
}
