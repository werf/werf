package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func GenerateInstallStage(dimgConfig interface{}, extra *builder.Extra) Interface {
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

func (s *InstallStage) Name() string {
	return "install"
}

func (s *InstallStage) GetContext(_ Cache) string {
	return s.builder.InstallChecksum() // TODO: git
}
