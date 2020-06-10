package stage

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

func GenerateFromStage(imageBaseConfig *config.StapelImageBase, baseImageRepoId string, baseStageOptions *NewBaseStageOptions) *FromStage {
	var baseImageRepoIdOrNone string
	if imageBaseConfig.FromLatest {
		baseImageRepoIdOrNone = baseImageRepoId
	}

	var fromImageOrArtifactImageName string
	if imageBaseConfig.FromImageName != "" {
		fromImageOrArtifactImageName = imageBaseConfig.FromImageName
	} else if imageBaseConfig.FromImageArtifactName != "" {
		fromImageOrArtifactImageName = imageBaseConfig.FromImageArtifactName
	}

	return newFromStage(fromImageOrArtifactImageName, baseImageRepoIdOrNone, imageBaseConfig.FromCacheVersion, baseStageOptions)
}

func newFromStage(fromImageOrArtifactImageName, baseImageRepoIdOrNone, cacheVersion string, baseStageOptions *NewBaseStageOptions) *FromStage {
	s := &FromStage{}
	s.cacheVersion = cacheVersion
	s.fromImageOrArtifactImageName = fromImageOrArtifactImageName
	s.baseImageRepoIdOrNone = baseImageRepoIdOrNone
	s.BaseStage = newBaseStage(From, baseStageOptions)
	return s
}

type FromStage struct {
	*BaseStage

	fromImageOrArtifactImageName string
	baseImageRepoIdOrNone        string
	cacheVersion                 string
}

func (s *FromStage) GetDependencies(c Conveyor, prevImage, _ container_runtime.ImageInterface) (string, error) {
	var args []string

	if s.cacheVersion != "" {
		args = append(args, s.cacheVersion)
	}

	if s.baseImageRepoIdOrNone != "" {
		args = append(args, s.baseImageRepoIdOrNone)
	}

	for _, mount := range s.configMounts {
		args = append(args, filepath.ToSlash(filepath.Clean(mount.From)), path.Clean(mount.To), mount.Type)
	}

	if s.fromImageOrArtifactImageName != "" {
		args = append(args, c.GetImageContentSignature(s.fromImageOrArtifactImageName))
	} else {
		args = append(args, prevImage.Name())
	}

	return util.Sha256Hash(args...), nil
}

func (s *FromStage) PrepareImage(_ Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
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
