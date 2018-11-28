package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func generateBeforeSetupStage(dimgConfig interface{}) Interface {
	b := getBuilder(dimgConfig)
	if b != nil && !b.IsBeforeSetupEmpty() {
		return NewBeforeSetupStage(b)
	}

	return nil
}

func NewBeforeSetupStage(builder builder.Builder) *BeforeSetupStage {
	s := &BeforeSetupStage{}
	s.UserStage = newUserStage(builder)
	return s
}

type BeforeSetupStage struct {
	*UserStage
}

func (s *BeforeSetupStage) Name() string {
	return "beforeSetup"
}

func (s *BeforeSetupStage) GetContext(_ Cache) string {
	return s.builder.BeforeSetupChecksum() // TODO: git
}
