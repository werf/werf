package stage

import (
	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
)

func GenerateBeforeInstallStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *BeforeInstallStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeInstallEmpty() {
		return newBeforeInstallStage(b, baseStageOptions)
	}

	return nil
}

func newBeforeInstallStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *BeforeInstallStage {
	s := &BeforeInstallStage{}
	s.UserStage = newUserStage(builder, BeforeInstall, baseStageOptions)
	return s
}

type BeforeInstallStage struct {
	*UserStage
}

func (s *BeforeInstallStage) GetDependencies(_ Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	return s.builder.BeforeInstallChecksum(), nil
}

func (s *BeforeInstallStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.BeforeInstall(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
