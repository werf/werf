package stage

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

func GenerateFromStage(imageBaseConfig *config.StapelImageBase, baseImageRepoId string, baseStageOptions *BaseStageOptions) *FromStage {
	var baseImageRepoIdOrNone string
	if imageBaseConfig.FromLatest {
		baseImageRepoIdOrNone = baseImageRepoId
	}

	var fromImageOrArtifactImageName string
	if imageBaseConfig.FromImageName != "" {
		fromImageOrArtifactImageName = imageBaseConfig.FromImageName
	} else if imageBaseConfig.FromArtifactName != "" {
		fromImageOrArtifactImageName = imageBaseConfig.FromArtifactName
	}

	return newFromStage(fromImageOrArtifactImageName, baseImageRepoIdOrNone, imageBaseConfig.FromCacheVersion, baseStageOptions)
}

func newFromStage(fromImageOrArtifactImageName, baseImageRepoIdOrNone, cacheVersion string, baseStageOptions *BaseStageOptions) *FromStage {
	s := &FromStage{}
	s.cacheVersion = cacheVersion
	s.fromImageOrArtifactImageName = fromImageOrArtifactImageName
	s.baseImageRepoIdOrNone = baseImageRepoIdOrNone
	s.BaseStage = NewBaseStage(From, baseStageOptions)
	return s
}

type FromStage struct {
	*BaseStage

	fromImageOrArtifactImageName string
	baseImageRepoIdOrNone        string
	cacheVersion                 string
}

func (s *FromStage) HasPrevStage() bool {
	return false
}

func (s *FromStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
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
		args = append(args, c.GetImageContentDigest(s.targetPlatform, s.fromImageOrArtifactImageName))
	} else {
		args = append(args, prevImage.Image.Name())
	}

	return util.Sha256Hash(args...), nil
}

func (s *FromStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	addLabels := map[string]string{imagePkg.WerfProjectRepoCommitLabel: c.GiterminismManager().HeadCommit()}
	if c.UseLegacyStapelBuilder(cb) {
		stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(addLabels)
	} else {
		stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
	}

	serviceMounts := s.getServiceMounts(prevBuiltImage)
	s.addServiceMountsLabels(serviceMounts, c, cb, stageImage)
	if !c.UseLegacyStapelBuilder(cb) {
		if err := s.addServiceMountsVolumes(serviceMounts, c, cb, stageImage, true); err != nil {
			return fmt.Errorf("error adding mounts volumes: %w", err)
		}
	}

	customMounts := s.getCustomMounts(prevBuiltImage)
	s.addCustomMountLabels(customMounts, c, cb, stageImage)
	if !c.UseLegacyStapelBuilder(cb) {
		if err := s.addCustomMountVolumes(customMounts, c, cb, stageImage, true); err != nil {
			return fmt.Errorf("error adding mounts volumes: %w", err)
		}
	}

	var mountpoints []string
	for _, mountCfg := range s.configMounts {
		mountpoints = append(mountpoints, mountCfg.To)
	}
	if len(mountpoints) > 0 {
		if c.UseLegacyStapelBuilder(cb) {
			mountpointsStr := strings.Join(mountpoints, " ")
			stageImage.Builder.LegacyStapelStageBuilder().Container().AddServiceRunCommands(fmt.Sprintf("%s -rf %s", stapel.RmBinPath(), mountpointsStr))
		}
	}

	return nil
}
