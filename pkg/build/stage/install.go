package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateInstallStage(dimgBaseConfig *config.DimgBase, extra *builder.Extra, baseStageOptions *NewBaseStageOptions) *InstallStage {
	b := getBuilder(dimgBaseConfig, extra)
	if b != nil && !b.IsInstallEmpty() {
		return newInstallStage(b, baseStageOptions)
	}

	return nil
}

func newInstallStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *InstallStage {
	s := &InstallStage{}
	s.UserWithGAPatchStage = newUserWithGAPatchStage(builder, Install, baseStageOptions)
	return s
}

type InstallStage struct {
	*UserWithGAPatchStage
}

func (s *InstallStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(Install)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.InstallChecksum(), stageDependenciesChecksum), nil
}

func (s *InstallStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.UserWithGAPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	if err := s.builder.Install(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
