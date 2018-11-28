package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func GenerateSetupStage(dimgConfig interface{}, extra *builder.Extra) Interface {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsSetupEmpty() {
		return newSetupStage(b)
	}

	return nil
}

func newSetupStage(builder builder.Builder) *SetupStage {
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
