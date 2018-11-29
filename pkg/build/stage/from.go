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

func (s *BaseFromStage) GetDependencies() (string, error) {
	var args []string

	args = append(args, s.cacheVersion)

	for _, mount := range s.mounts {
		args = append(args, mount.From, mount.To, mount.Type)
	}

	return util.Sha256Hash(args...), nil
}

type FromStage struct {
	*BaseFromStage

	from string
}

func (s *FromStage) GetDependencies(_ Conveyor, _ Image) (string, error) {
	baseFromStageDependencies, err := s.BaseFromStage.GetDependencies()
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(baseFromStageDependencies, s.from), nil
}

type FromDimgStage struct {
	*BaseFromStage

	dimgName string
}

func (s *FromDimgStage) GetDependencies(c Conveyor, _ Image) (string, error) {
	baseFromStageDependencies, err := s.BaseFromStage.GetDependencies()
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(baseFromStageDependencies, c.GetDimgSignature(s.dimgName)), nil
}
