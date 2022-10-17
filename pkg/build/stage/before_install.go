package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
)

func GenerateBeforeInstallStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, baseStageOptions *BaseStageOptions) *BeforeInstallStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeInstallEmpty(ctx) {
		return newBeforeInstallStage(b, baseStageOptions)
	}

	return nil
}

func newBeforeInstallStage(builder builder.Builder, baseStageOptions *BaseStageOptions) *BeforeInstallStage {
	s := &BeforeInstallStage{}
	s.UserStage = newUserStage(builder, BeforeInstall, baseStageOptions)
	return s
}

type BeforeInstallStage struct {
	*UserStage
}

func (s *BeforeInstallStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	return s.builder.BeforeInstallChecksum(ctx), nil
}

func (s *BeforeInstallStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := s.BaseStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil); err != nil {
		return err
	}

	if err := s.builder.BeforeInstall(ctx, cb, stageImage.Builder, c.UseLegacyStapelBuilder(cb)); err != nil {
		return err
	}

	return nil
}
