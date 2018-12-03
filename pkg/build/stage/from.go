package stage

import (
	"fmt"
	"strings"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateFromStage(dimgBaseConfig *config.DimgBase, baseStageOptions *NewBaseStageOptions) *FromStage {
	return newFromStage(dimgBaseConfig.FromCacheVersion, dimgBaseConfig.Mount, baseStageOptions)
}

func newFromStage(cacheVersion string, mounts []*config.Mount, baseStageOptions *NewBaseStageOptions) *FromStage {
	s := &FromStage{}
	s.cacheVersion = cacheVersion
	s.mounts = mounts
	s.BaseStage = newBaseStage(baseStageOptions)
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

func (s *FromStage) GetDependencies(_ Conveyor, image image.Image) (string, error) {
	var args []string

	args = append(args, s.cacheVersion)

	for _, mount := range s.mounts {
		args = append(args, mount.From, mount.To, mount.Type)
	}

	args = append(args, image.Name())

	return util.Sha256Hash(args...), nil
}

func (s *FromStage) PrepareImage(prevBuiltImage, image image.Image) error {
	var err error

	err = s.addServiceMounts(prevBuiltImage, image, true)
	if err != nil {
		return fmt.Errorf("error adding service mounts: %s", err)
	}

	err = s.addCustomMounts(prevBuiltImage, image, true)
	if err != nil {
		return fmt.Errorf("error adding custom mounts: %s", err)
	}

	mountpoints := []string{}
	for _, mountCfg := range s.mounts {
		mountpoints = append(mountpoints, mountCfg.To)
	}
	if len(mountpoints) != 0 {
		mountpointsStr := strings.Join(mountpoints, " ")
		image.Container().AddServiceRunCommands(fmt.Sprintf("%s -rf %s", dappdeps.RmBinPath(), mountpointsStr))
	}

	return nil
}
