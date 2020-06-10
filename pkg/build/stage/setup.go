package stage

import (
	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/util"
)

func GenerateSetupStage(imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *SetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsSetupEmpty() {
		return newSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *SetupStage {
	s := &SetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Setup, gitPatchStageOptions, baseStageOptions)
	return s
}

type SetupStage struct {
	*UserWithGitPatchStage
}

func (s *SetupStage) GetDependencies(c Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(c, Setup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.SetupChecksum(), stageDependenciesChecksum), nil
}

func (s *SetupStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.Setup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
