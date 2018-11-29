package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func GenerateFromStage(dimgBaseConfig *config.DimgBase) Interface {
	return newFromStage(dimgBaseConfig.FromCacheVersion, dimgBaseConfig.Mount)
}

func newFromStage(cacheVersion string, mounts []*config.Mount) *FromStage {
	s := &FromStage{}
	s.cacheVersion = cacheVersion
	s.mounts = mounts
	s.BaseStage = newBaseStage()

	return s
}

type FromStage struct {
	*BaseStage

	cacheVersion string
	mounts       []*config.Mount
}

func (s *FromStage) Name() StageName {
	return From
}

func (s *FromStage) GetDependencies(_ Conveyor, baseImage Image) (string, error) {
	var args []string

	args = append(args, s.cacheVersion)

	for _, mount := range s.mounts {
		args = append(args, mount.From, mount.To, mount.Type)
	}

	args = append(args, baseImage.GetName())

	return util.Sha256Hash(args...), nil
}
