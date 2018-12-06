package stage

import (
	"fmt"
	"path/filepath"
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
	s.BaseStage = newBaseStage(From, baseStageOptions)
	return s
}

type FromStage struct {
	*BaseStage

	cacheVersion string
	mounts       []*config.Mount
}

func (s *FromStage) GetDependencies(_ Conveyor, prevImage image.Image) (string, error) {
	var args []string

	if s.cacheVersion != "" {
		args = append(args, s.cacheVersion)
	}

	for _, mount := range s.mounts {
		args = append(args, mount.From, mount.To, mount.Type)
	}

	args = append(args, prevImage.Name())

	return util.Sha256Hash(args...), nil
}

func (s *FromStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	serviceMounts := mergeMounts(s.getServiceMountsFromConfig(), s.getServiceMountsFromLabels(prevBuiltImage))
	s.addServiceMountsLabels(serviceMounts, image)

	customMounts := mergeMounts(s.getCustomMountsFromConfig(), s.getCustomMountsFromLabels(prevBuiltImage))
	s.addCustomMountLabels(customMounts, image)

	s.addCleanupMountsCommands(image)

	return nil
}

func (s *FromStage) getServiceMountsFromConfig() map[string][]string {
	mountpointsByType := map[string][]string{}

	for _, mountCfg := range s.mounts {
		mountpoint := filepath.Clean(mountCfg.To)
		mountpointsByType[mountCfg.Type] = append(mountpointsByType[mountCfg.Type], mountpoint)
	}

	return mountpointsByType
}

func (s *FromStage) getCustomMountsFromConfig() map[string][]string {
	mountpointsByFrom := map[string][]string{}
	for _, mountCfg := range s.mounts {
		if mountCfg.Type != "custom_dir" {
			continue
		}

		from := filepath.Clean(mountCfg.From)
		mountpoint := filepath.Clean(mountCfg.To)

		mountpointsByFrom[from] = util.UniqAppendString(mountpointsByFrom[from], mountpoint)
	}

	return mountpointsByFrom
}

func (s *FromStage) addCleanupMountsCommands(image image.Image) {
	var mountpoints []string

	for _, mountCfg := range s.mounts {
		mountpoints = append(mountpoints, mountCfg.To)
	}

	if len(mountpoints) != 0 {
		mountpointsStr := strings.Join(mountpoints, " ")
		image.Container().AddServiceRunCommands(fmt.Sprintf("%s -rf %s", dappdeps.RmBinPath(), mountpointsStr))
	}
}

func mergeMounts(mountsFromConfig, mountsFromLabels map[string][]string) map[string][]string {
	resultMounts := map[string][]string{}

	var keys []string
	for key := range mountsFromConfig {
		keys = append(keys, key)
	}

	for key := range mountsFromLabels {
		keys = append(keys, key)
	}

	isNotInArray := func(arr []string, elm string) bool {
		for _, arrElm := range arr {
			if arrElm == elm {
				return false
			}
		}

		return true
	}

	for _, key := range keys {
		mountsByKeyFromConfig, ok := mountsFromConfig[key]
		if ok {
			resultMounts[key] = mountsByKeyFromConfig
		} else {
			resultMounts[key] = mountsFromConfig[key]
			continue
		}

		mountsByKeyFromLabels, ok := mountsFromLabels[key]
		if ok {
			for _, mount := range mountsByKeyFromLabels {
				if isNotInArray(mountsByKeyFromConfig, mount) {
					resultMounts[key] = append(resultMounts[key], mount)
				}
			}
		}
	}

	return resultMounts
}
