package stage

import (
	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/util"
)

func GenerateInstallStage(imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *InstallStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
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

func (s *InstallStage) GetDependencies(c Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(c, Install)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.InstallChecksum(), stageDependenciesChecksum), nil
}

func (s *InstallStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.Install(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
