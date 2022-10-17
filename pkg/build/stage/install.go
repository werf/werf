package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/util"
)

func GenerateInstallStage(ctx context.Context, imageBaseConfig *config.StapelImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *InstallStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsInstallEmpty(ctx) {
		return newInstallStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newInstallStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *BaseStageOptions) *InstallStage {
	s := &InstallStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Install, gitPatchStageOptions, baseStageOptions)
	return s
}

type InstallStage struct {
	*UserWithGitPatchStage
}

func (s *InstallStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(ctx, c, Install)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.InstallChecksum(ctx), stageDependenciesChecksum), nil
}

func (s *InstallStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := s.UserWithGitPatchStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil); err != nil {
		return err
	}

	if err := s.builder.Install(ctx, cb, stageImage.Builder, c.UseLegacyStapelBuilder(cb)); err != nil {
		return err
	}

	return nil
}
