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

func GenerateSetupStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *SetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsSetupEmpty(ctx) {
		return newSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *SetupStage {
	s := &SetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Setup, gitPatchStageOptions, baseStageOptions)
	return s
}

type SetupStage struct {
	*UserWithGitPatchStage
}

func (s *SetupStage) GetDependencies(ctx context.Context, c Conveyor, prevBuiltImage, _ container_runtime.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(ctx, c, Setup)
	if err != nil {
		return "", err
	}

	var devModeChecksum string
	if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[imagePkg.WerfDevLabel] != "true" {
		devModeChecksum, err = s.getStageDependenciesStagingStatusChecksum(ctx, c, Setup)
		if err != nil {
			return "", err
		}
	}

	var args []string
	args = append(args, s.builder.SetupChecksum(ctx))
	args = append(args, stageDependenciesChecksum)

	if devModeChecksum != "" {
		args = append(args, devModeChecksum)
	}

	return util.Sha256Hash(args...), nil
}

func (s *SetupStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.Setup(ctx, image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
