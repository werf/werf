package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/util"
)

func GenerateBeforeSetupStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *BeforeSetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeSetupEmpty(ctx) {
		return newBeforeSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newBeforeSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *BeforeSetupStage {
	s := &BeforeSetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, BeforeSetup, gitPatchStageOptions, baseStageOptions)
	return s
}

type BeforeSetupStage struct {
	*UserWithGitPatchStage
}

func (s *BeforeSetupStage) GetDependencies(ctx context.Context, c Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(ctx, c, BeforeSetup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.BeforeSetupChecksum(ctx), stageDependenciesChecksum), nil
}

func (s *BeforeSetupStage) PrepareImage(ctx context.Context, c Conveyor, cr container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage) error {
	if err := s.UserWithGitPatchStage.PrepareImage(ctx, c, cr, prevBuiltImage, stageImage); err != nil {
		return err
	}

	if err := s.builder.BeforeSetup(ctx, cr, stageImage.Builder, c.UseLegacyStapelBuilder(cr)); err != nil {
		return err
	}

	return nil
}
