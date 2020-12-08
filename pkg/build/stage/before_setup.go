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

func GenerateBeforeSetupStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeSetupEmpty(ctx) {
		return newBeforeSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newBeforeSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	s := &BeforeSetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, BeforeSetup, gitPatchStageOptions, baseStageOptions)
	return s
}

type BeforeSetupStage struct {
	*UserWithGitPatchStage
}

func (s *BeforeSetupStage) GetDependencies(ctx context.Context, c Conveyor, prevBuiltImage, _ container_runtime.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(ctx, c, BeforeSetup)
	if err != nil {
		return "", err
	}

	var devModeChecksum string
	if giterminism_inspector.DevMode && prevBuiltImage.GetStageDescription().Info.Labels[imagePkg.WerfDevLabel] != "true" {
		devModeChecksum, err = s.getStageDependenciesStagingStatusChecksum(ctx, c, BeforeSetup)
		if err != nil {
			return "", err
		}
	}

	var args []string
	args = append(args, s.builder.BeforeSetupChecksum(ctx))
	args = append(args, stageDependenciesChecksum)

	if devModeChecksum != "" {
		args = append(args, devModeChecksum)
	}

	return util.Sha256Hash(args...), nil
}

func (s *BeforeSetupStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.BeforeSetup(ctx, image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
