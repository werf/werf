package stage

import (
	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/util"
)

func GenerateBeforeSetupStage(imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeSetupEmpty() {
		return newBeforeSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newBeforeSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	s := &BeforeSetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, BeforeSetup, gitPatchStageOptions, baseStageOptions)
	return s
}

type BeforeSetupStage struct {
	*UserWithGitPatchStage
}

func (s *BeforeSetupStage) GetDependencies(c Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(c, BeforeSetup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.BeforeSetupChecksum(), stageDependenciesChecksum), nil
}

func (s *BeforeSetupStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.BeforeSetup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
