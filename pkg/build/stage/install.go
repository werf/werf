package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
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

func (s *InstallStage) GetContext(_ Cache) string {
	return util.Sha256Hash(
		s.builder.InstallChecksum(),
		s.GetStageDependenciesChecksum(Install),
	)
}
