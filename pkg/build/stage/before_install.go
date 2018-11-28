package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func generateBeforeInstallStage(dimgConfig interface{}) Interface {
	b := getBuilder(dimgConfig)
	if b != nil && !b.IsBeforeInstallEmpty() {
		return NewBeforeInstallStage(b)
	}

	return nil
}

func NewBeforeInstallStage(builder builder.Builder) *BeforeInstallStage {
	s := &BeforeInstallStage{}
	s.UserStage = newUserStage(builder)
	return s
}

type BeforeInstallStage struct {
	*UserStage
}

func (s *BeforeInstallStage) Name() string {
	return "beforeInstall"
}

func (s *BeforeInstallStage) GetContext(_ Cache) string {
	return s.builder.BeforeInstallChecksum() // TODO: git
}
