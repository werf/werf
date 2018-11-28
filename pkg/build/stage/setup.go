package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func generateSetupStage(dimgConfig interface{}) Interface {
	b := getBuilder(dimgConfig)
	if b != nil && !b.IsSetupEmpty() {
		return NewSetupStage(b)
	}

	return nil
}

func NewSetupStage(builder builder.Builder) *SetupStage {
	s := &SetupStage{}
	s.UserStage = newUserStage(builder)
	return s
}

type SetupStage struct {
	*UserStage
}

func (s *SetupStage) Name() string {
	return "setup"
}

func (s *SetupStage) GetContext(_ Cache) string {
	return s.builder.SetupChecksum() // TODO: git
}
