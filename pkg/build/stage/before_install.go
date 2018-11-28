package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
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

func (s *BeforeInstallStage) GetDependencies(_ Cache) string {
	return s.builder.BeforeInstallChecksum()
}

func (s *BeforeInstallStage) GetContext(_ Cache) string {
	return ""
}
