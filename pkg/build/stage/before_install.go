package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
)

func GenerateBeforeInstallStage(dimgConfig config.DimgInterface, extra *builder.Extra) Interface {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsBeforeInstallEmpty() {
		return newBeforeInstallStage(b)
	}

	return nil
}

func newBeforeInstallStage(builder builder.Builder) *BeforeInstallStage {
	s := &BeforeInstallStage{}
	s.UserStage = newUserStage(builder)
	return s
}

type BeforeInstallStage struct {
	*UserStage
}

func (s *BeforeInstallStage) Name() StageName {
	return BeforeInstall
}

func (s *BeforeInstallStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	return s.builder.BeforeInstallChecksum(), nil
}

func (s *BeforeInstallStage) GetContext(_ Conveyor) (string, error) {
	return "", nil
}

func (s *BeforeInstallStage) PrepareImage(prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.BeforeInstall(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
