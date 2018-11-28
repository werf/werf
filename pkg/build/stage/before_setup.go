package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func GenerateBeforeSetupStage(dimgConfig interface{}, extra *builder.Extra) Interface {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsBeforeSetupEmpty() {
		return newBeforeSetupStage(b)
	}

	return nil
}

func newBeforeSetupStage(builder builder.Builder) *BeforeSetupStage {
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
