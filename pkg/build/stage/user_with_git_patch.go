package stage

import (
	"fmt"

	"github.com/flant/werf/pkg/storage"

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

func (s *UserWithGitPatchStage) SelectCacheImage(images []*storage.ImageInfo) (*storage.ImageInfo, error) {
	ancestorsImages, err := s.selectCacheImagesAncestorsByGitMappings(images)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectCacheImageByOldestCreationTimestamp(ancestorsImages)
}

func (s *UserWithGitPatchStage) GetNextStageDependencies(c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(c)
}

func (s *UserWithGitPatchStage) IsEmpty(c Conveyor, prevBuiltImage image.ImageInterface) (bool, error) {
	if empty, err := s.UserStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	return s.GitPatchStage.IsEmpty(c, prevBuiltImage)
}

func (s *UserWithGitPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.GitPatchStage.prepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	return nil
}
