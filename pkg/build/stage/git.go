package stage

import (
	"context"

	"github.com/werf/werf/pkg/container_backend"
	imagePkg "github.com/werf/werf/pkg/image"
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
	if err := s.BaseStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil); err != nil {
		return err
	}

	if c.GiterminismManager().Dev() {
		addLabels := map[string]string{imagePkg.WerfDevLabel: "true"}
		if c.UseLegacyStapelBuilder(cb) {
			stageImage.Builder.LegacyStapelStageBuilder().BuilderContainer().AddLabel(addLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
		}
	}

	return nil
}
