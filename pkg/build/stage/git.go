package stage

import (
	"context"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

func newGitStage(name StageName, baseStageOptions *BaseStageOptions) *GitStage {
	s := &GitStage{}
	s.BaseStage = NewBaseStage(name, baseStageOptions)
	return s
}

type GitStage struct {
	*BaseStage
}

func (s *GitStage) IsEmpty(ctx context.Context, _ Conveyor, _ *StageImage) (bool, error) {
	return s.isEmpty(ctx), nil
}

func (s *GitStage) isEmpty(_ context.Context) bool {
	return len(s.gitMappings) == 0
}

func (s *GitStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	return s.BaseStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil)
}

func (s *GitStage) SelectSuitableStageDesc(ctx context.Context, c Conveyor, stageDescSet image.StageDescSet) (*image.StageDesc, error) {
	return s.selectAncestorStageDescByGitMappings(ctx, c, stageDescSet)
}
