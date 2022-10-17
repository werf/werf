package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/util"
)

func GenerateSetupStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *SetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsSetupEmpty(ctx) {
		return newSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *SetupStage {
	s := &SetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Setup, gitPatchStageOptions, baseStageOptions)
	return s
}

type SetupStage struct {
	*UserWithGitPatchStage
}

func (s *SetupStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(ctx, c, Setup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.SetupChecksum(ctx), stageDependenciesChecksum), nil
}

func (s *SetupStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := s.UserWithGitPatchStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil); err != nil {
		return err
	}

	if err := s.builder.Setup(ctx, cb, stageImage.Builder, c.UseLegacyStapelBuilder(cb)); err != nil {
		return err
	}

	return nil
}
