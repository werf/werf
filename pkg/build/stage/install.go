package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func generateInstallStage(dimgConfig interface{}) Interface {
	b := getBuilder(dimgConfig)
	if b != nil && !b.IsInstallEmpty() {
		return NewInstallStage(b)
	}

	return nil
}

func NewInstallStage(builder builder.Builder) *InstallStage {
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
