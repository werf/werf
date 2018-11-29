package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func GenerateSetupStage(dimgConfig config.DimgInterface, extra *builder.Extra) Interface {
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

func (s *SetupStage) Name() StageName {
	return Setup
}

func (s *SetupStage) GetContext(_ Conveyor) string {
	return util.Sha256Hash(
		s.builder.SetupChecksum(),
		s.GetStageDependenciesChecksum(Setup),
	)
}
