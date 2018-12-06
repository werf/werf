package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func getBuilder(dimgBaseConfig *config.DimgBase, extra *builder.Extra) builder.Builder {
	var b builder.Builder
	if dimgBaseConfig.Shell != nil {
		b = builder.NewShellBuilder(dimgBaseConfig.Shell)
	} else if dimgBaseConfig.Ansible != nil {
		b = builder.NewAnsibleBuilder(dimgBaseConfig.Ansible, extra)
	}

	return b
}

func newUserStage(builder builder.Builder, name StageName, baseStageOptions *NewBaseStageOptions) *UserStage {
	s := &UserStage{}
	s.builder = builder
	s.BaseStage = newBaseStage(name, baseStageOptions)
	return s
}

type UserStage struct {
	*BaseStage

	builder builder.Builder
}

func (s *UserStage) getStageDependenciesChecksum(name StageName) (string, error) {
	var args []string
	for _, ga := range s.gitArtifacts {
		checksum, err := ga.StageDependenciesChecksum(name)
		if err != nil {
			return "", err
		}

		args = append(args, checksum)
	}

	return util.Sha256Hash(args...), nil
}
