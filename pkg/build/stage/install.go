package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_inspector"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

func GenerateInstallStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *InstallStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsInstallEmpty(ctx) {
		return newInstallStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newInstallStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *InstallStage {
	s := &InstallStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Install, gitPatchStageOptions, baseStageOptions)
	return s
}

type InstallStage struct {
	*UserWithGitPatchStage
}

func (s *InstallStage) GetDependencies(ctx context.Context, c Conveyor, prevBuiltImage, _ container_runtime.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(ctx, c, Install)
	if err != nil {
		return "", err
	}

	var devModeChecksum string
	if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[imagePkg.WerfDevLabel] != "true" {
		devModeChecksum, err = s.getStageDependenciesStagingStatusChecksum(ctx, c, Install)
		if err != nil {
			return "", err
		}
	}

	var args []string
	args = append(args, s.builder.InstallChecksum(ctx))
	args = append(args, stageDependenciesChecksum)

	if devModeChecksum != "" {
		args = append(args, devModeChecksum)
	}

	return util.Sha256Hash(args...), nil
}

func (s *InstallStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.Install(ctx, image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
