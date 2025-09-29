package stage

import (
	"context"

	"github.com/werf/werf/v2/pkg/build/builder"
	"github.com/werf/werf/v2/pkg/container_backend"
)

func newUserWithGitPatchStage(builder builder.Builder, name StageName, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *UserWithGitPatchStage {
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

func (s *UserWithGitPatchStage) GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(ctx, c)
}

func (s *UserWithGitPatchStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := s.BaseStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil); err != nil {
		return err
	}

	if isPatchEmpty, err := s.GitPatchStage.IsEmpty(ctx, c, prevBuiltImage); err != nil {
		return err
	} else if !isPatchEmpty {
		if err := s.GitPatchStage.prepareImage(ctx, c, cb, prevBuiltImage, stageImage); err != nil {
			return err
		}
	}

	return nil
}
