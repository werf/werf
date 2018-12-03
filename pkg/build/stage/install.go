package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateInstallStage(dimgConfig config.DimgInterface, extra *builder.Extra) Interface {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsInstallEmpty() {
		return newInstallStage(b)
	}

	return nil
}

func newInstallStage(builder builder.Builder) *InstallStage {
	s := &InstallStage{}
	s.UserStage = newUserStage(builder)
	return s
}

type InstallStage struct {
	*UserStage
}

func (s *InstallStage) Name() StageName {
	return Install
}

func (s *InstallStage) GetContext(_ Conveyor) (string, error) {
	stageDependenciesChecksum, err := s.GetStageDependenciesChecksum(Install)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.InstallChecksum(), stageDependenciesChecksum), nil
}

func (s *InstallStage) PrepareImage(prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.Install(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
