package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func generateFromStage(dimgBaseConfig *config.DimgBase) Interface {
	fromCacheVersion := dimgBaseConfig.FromCacheVersion

	if dimgBaseConfig.From != "" {
		return newFromStage(fromCacheVersion, dimgBaseConfig.From)
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
			return newFromDimgStage(fromCacheVersion, fromDimgName)
		}
	}

	return nil
}

func newFromStage(cacheVersion, from string) *FromStage {
	s := &FromStage{}
	s.from = from
	s.BaseFromStage = newBaseFromStage(cacheVersion)

	return s
}

func newFromDimgStage(cacheVersion, dimgName string) *FromDimgStage {
	s := &FromDimgStage{}
	s.dimgName = dimgName
	s.BaseFromStage = newBaseFromStage(cacheVersion)

	return s
}

func newBaseFromStage(cacheVersion string) *BaseFromStage {
	s := &BaseFromStage{}
	s.cacheVersion = cacheVersion
	s.BaseStage = newBaseStage()

	return s
}

type BaseFromStage struct { // TODO: mounts
	*BaseStage

	cacheVersion string
}

func (s *BaseFromStage) Name() string {
	return "name"
}

func (s *BaseFromStage) GetDependencies() string {
	return util.Sha256Hash(s.cacheVersion)
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
