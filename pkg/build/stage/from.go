package stage

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/util"
)

func GenerateFromStage(imageBaseConfig *config.StapelImageBase, baseImageRepoId string, baseStageOptions *NewBaseStageOptions) *FromStage {
	var baseImageRepoIdOrNone string
	if imageBaseConfig.FromLatest {
		baseImageRepoIdOrNone = baseImageRepoId
	}

	return newFromStage(baseImageRepoIdOrNone, imageBaseConfig.FromCacheVersion, baseStageOptions)
}

func newFromStage(baseImageRepoIdOrNone, cacheVersion string, baseStageOptions *NewBaseStageOptions) *FromStage {
	s := &FromStage{}
	s.cacheVersion = cacheVersion
	s.baseImageRepoIdOrNone = baseImageRepoIdOrNone
	s.BaseStage = newBaseStage(From, baseStageOptions)
	return s
}

type FromStage struct {
	*BaseStage

	baseImageRepoIdOrNone string
	cacheVersion          string
}

func (s *FromStage) GetDependencies(_ Conveyor, prevImage, _ image.ImageInterface) (string, error) {
	var args []string

	if s.cacheVersion != "" {
		args = append(args, s.cacheVersion)
	}

	if s.baseImageRepoIdOrNone != "" {
		args = append(args, s.baseImageRepoIdOrNone)
	}

	for _, mount := range s.configMounts {
		args = append(args, filepath.Clean(mount.From), path.Clean(mount.To), mount.Type)
	}

	args = append(args, prevImage.Name())

	return util.Sha256Hash(args...), nil
}

func (s *FromStage) PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	serviceMounts := s.getServiceMounts(prevBuiltImage)
	s.addServiceMountsLabels(serviceMounts, image)

	customMounts := s.getCustomMounts(prevBuiltImage)
	s.addCustomMountLabels(customMounts, image)

	var mountpoints []string
	for _, mountCfg := range s.configMounts {
		mountpoints = append(mountpoints, mountCfg.To)
	}
	if len(mountpoints) != 0 {
		mountpointsStr := strings.Join(mountpoints, " ")
		image.Container().AddServiceRunCommands(fmt.Sprintf("%s -rf %s", stapel.RmBinPath(), mountpointsStr))
	}

	return nil
}
