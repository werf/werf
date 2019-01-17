package stage

import (
	"github.com/flant/werf/pkg/build/builder"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

func GenerateInstallStage(dimgBaseConfig *config.DimgBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *InstallStage {
	b := getBuilder(dimgBaseConfig, baseStageOptions)
	if b != nil && !b.IsInstallEmpty() {
		return newInstallStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newInstallStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *InstallStage {
	s := &InstallStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Install, gitPatchStageOptions, baseStageOptions)
	return s
}

type InstallStage struct {
	*UserWithGitPatchStage
}

func (s *InstallStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(Install)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.InstallChecksum(), stageDependenciesChecksum), nil
}

func (s *InstallStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.UserWithGitPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	if err := s.builder.Install(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
