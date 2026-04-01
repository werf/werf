package stage

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/pkg/util/option"
)

func GenerateFromStage(imageBaseConfig *config.StapelImageBase, baseImageRepoId, imageCacheVersion string, baseStageOptions *BaseStageOptions) *FromStage {
	var baseImageRepoIdOrNone string
	if imageBaseConfig.FromLatest {
		baseImageRepoIdOrNone = baseImageRepoId
	}

	fromImageOrArtifactImageName := option.ValueOrDefault(imageBaseConfig.From, imageBaseConfig.FromArtifactName)

	s := &FromStage{}
	s.fromCacheVersion = imageBaseConfig.FromCacheVersion
	s.fromImageOrArtifactImageName = fromImageOrArtifactImageName
	s.baseImageRepoIdOrNone = baseImageRepoIdOrNone
	s.BaseStage = NewBaseStage(From, baseStageOptions)
	s.imageCacheVersion = imageCacheVersion
	s.fromExternal = imageBaseConfig.FromExternal
	s.fromScratch = imageBaseConfig.From == "scratch"
	return s
}

type FromStage struct {
	*BaseStage

	baseImageRepoIdOrNone        string
	fromCacheVersion             string
	fromImageOrArtifactImageName string
	fromExternal                 bool
	fromScratch                  bool

	imageCacheVersion string
}

func (s *FromStage) HasPrevStage() bool {
	return false
}

func (s *FromStage) IsBuildable() bool {
	// Scratch has no base filesystem to build from here: the first stage image is created later
	// in MutateImage via storage.PostManifest + config mutation.
	if s.fromScratch {
		return false
	}

	return s.BaseStage.IsBuildable()
}

func (s *FromStage) IsMutable() bool {
	return s.fromScratch
}

func (s *FromStage) GetDependencies(_ context.Context, c Conveyor, _ container_backend.ContainerBackend, prevImage, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	if s.imageCacheVersion != "" {
		args = append(args, s.imageCacheVersion)
	}

	if s.fromCacheVersion != "" {
		args = append(args, s.fromCacheVersion)
	}

	if s.baseImageRepoIdOrNone != "" {
		args = append(args, s.baseImageRepoIdOrNone)
	}

	for _, mount := range s.configMounts {
		args = append(args, filepath.ToSlash(filepath.Clean(mount.From)), path.Clean(mount.To), mount.Type)
	}

	if s.fromScratch {
		args = append(args, "scratch")
	} else if s.fromImageOrArtifactImageName != "" && !s.fromExternal {
		args = append(args, c.GetImageContentDigest(s.targetPlatform, s.fromImageOrArtifactImageName))
	} else {
		args = append(args, prevImage.Image.Name())
	}

	return util.Sha256Hash(args...), nil
}

func (s *FromStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, _ container_backend.BuildContextArchiver) error {
	addLabels := map[string]string{imagePkg.WerfProjectRepoCommitLabel: c.GiterminismManager().HeadCommit(ctx)}
	if c.UseLegacyStapelBuilder(cb) {
		stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(addLabels)
	} else {
		stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
	}

	// For scratch there is no parent image data to inherit mounts from, so only the build labels
	// are prepared here. The actual empty image is created later in MutateImage.
	if s.fromScratch {
		return nil
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

type ScratchImageStageStorage interface {
	ImageMutatorPusher
	PostManifest(ctx context.Context, ref string, opts container_backend.PostManifestOpts) error
}

func (s *FromStage) MutateImage(ctx context.Context, storage ImageMutatorPusher, prevBuiltImage, stageImage *StageImage) error {
	if !s.fromScratch {
		return s.BaseStage.MutateImage(ctx, storage, prevBuiltImage, stageImage)
	}

	scratchStorage, ok := storage.(ScratchImageStageStorage)
	if !ok {
		return fmt.Errorf("scratch from stage storage does not support manifest creation")
	}

	serviceLabels := stageImage.Image.GetBuildServiceLabels()
	labelSpecs := make([]string, 0, len(serviceLabels))
	for key, value := range serviceLabels {
		labelSpecs = append(labelSpecs, fmt.Sprintf("%s=%s", key, value))
	}

	// Scratch starts from an empty image. First create the manifest with service labels, then mutate
	// the resulting image config so downstream stages see the same metadata as on regular stage images.
	if err := scratchStorage.PostManifest(ctx, stageImage.Image.Name(), container_backend.PostManifestOpts{Labels: labelSpecs}); err != nil {
		return fmt.Errorf("create scratch stage image: %w", err)
	}

	return scratchStorage.MutateAndPushImage(ctx, stageImage.Image.Name(), stageImage.Image.Name(), imagePkg.SpecConfig{Labels: serviceLabels}, stageImage.Image)
}
