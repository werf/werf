package stage

import (
	"github.com/flant/werf/pkg/build/builder"
	"github.com/flant/werf/pkg/image"
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

func (s *UserWithGitPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	stageName := c.GetBuildingGitStage(s.dimgName)
	if stageName == s.Name() {
		if err := s.GitPatchStage.prepareImage(c, prevBuiltImage, image); err != nil {
			return nil
		}
	}

	return nil
}

func (s *UserWithGitPatchStage) AfterImageSyncDockerStateHook(c Conveyor) error {
	if err := s.GitPatchStage.AfterImageSyncDockerStateHook(c); err != nil {
		return err
	}

	return nil
}
