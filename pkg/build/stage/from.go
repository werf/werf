package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func GenerateFromStage(dimgBaseConfig *config.DimgBase) Interface {
	fromCacheVersion := dimgBaseConfig.FromCacheVersion

	if dimgBaseConfig.From != "" {
		return newFromStage(fromCacheVersion, dimgBaseConfig.Mount, dimgBaseConfig.From)
	} else {
		fromDimg := dimgBaseConfig.FromDimg
		fromDimgArtifact := dimgBaseConfig.FromDimgArtifact

		var fromDimgName string
		if fromDimg != nil {
			fromDimgName = fromDimg.Name
		} else {
			fromDimgName = fromDimgArtifact.Name
		}

		if fromDimgName != "" {
			return newFromDimgStage(fromCacheVersion, dimgBaseConfig.Mount, fromDimgName)
		}
	}

	return nil
}

func newFromStage(cacheVersion string, mounts []*config.Mount, from string) *FromStage {
	s := &FromStage{}
	s.from = from
	s.BaseFromStage = newBaseFromStage(cacheVersion, mounts)

	return s
}

func newFromDimgStage(cacheVersion string, mounts []*config.Mount, dimgName string) *FromDimgStage {
	s := &FromDimgStage{}
	s.dimgName = dimgName
	s.BaseFromStage = newBaseFromStage(cacheVersion, mounts)

	return s
}

func newBaseFromStage(cacheVersion string, mounts []*config.Mount) *BaseFromStage {
	s := &BaseFromStage{}
	s.cacheVersion = cacheVersion
	s.mounts = mounts
	s.BaseStage = newBaseStage()

	return s
}

type BaseFromStage struct {
	*BaseStage

	cacheVersion string
	mounts       []*config.Mount
}

func (s *BaseFromStage) Name() StageName {
	return From
}

func (s *BaseFromStage) GetDependencies() string {
	var args []string

	args = append(args, s.cacheVersion)

	for _, mount := range s.mounts {
		args = append(args, mount.From, mount.To, mount.Type)
	}

	return util.Sha256Hash(args...)
}

type FromStage struct {
	*BaseFromStage

	from string
}

func (s *FromStage) GetDependencies(_ Cache) string {
	return util.Sha256Hash(s.BaseFromStage.GetDependencies(), s.from)
}

type FromDimgStage struct {
	*BaseFromStage

	dimgName string
}

func (s *FromDimgStage) GetDependencies(c Cache) string {
	return util.Sha256Hash(
		s.BaseFromStage.GetDependencies(),
		c.GetDimg(s.dimgName).LatestStage().GetSignature(),
	)
}
