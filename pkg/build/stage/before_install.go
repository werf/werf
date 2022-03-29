package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
)

func GenerateBeforeInstallStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *BeforeInstallStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeInstallEmpty(ctx) {
		return newBeforeInstallStage(b, baseStageOptions)
	}

	return nil
}

func newBeforeInstallStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *BeforeInstallStage {
	s := &BeforeInstallStage{}
	s.UserStage = newUserStage(builder, BeforeInstall, baseStageOptions)
	return s
}

type BeforeInstallStage struct {
	*UserStage
}

func (s *BeforeInstallStage) GetDependencies(ctx context.Context, _ Conveyor, _, _ *StageImage) (string, error) {
	return s.builder.BeforeInstallChecksum(ctx), nil
}

func (s *BeforeInstallStage) PrepareImage(ctx context.Context, c Conveyor, cr container_runtime.ContainerRuntime, prevBuiltImage, stageImage *StageImage) error {
	if err := s.BaseStage.PrepareImage(ctx, c, cr, prevBuiltImage, stageImage); err != nil {
		return err
	}

	if err := s.builder.BeforeInstall(ctx, cr, stageImage.Builder, c.UseLegacyStapelBuilder(cr)); err != nil {
		return err
	}

	return nil
}
