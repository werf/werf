package stage

import (
	"context"

	"github.com/werf/werf/pkg/container_backend"
	imagePkg "github.com/werf/werf/pkg/image"
)

func newGitStage(name StageName, baseStageOptions *NewBaseStageOptions) *GitStage {
	s := &GitStage{}
	s.BaseStage = newBaseStage(name, baseStageOptions)
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

func (s *GitStage) PrepareImage(ctx context.Context, c Conveyor, cr container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage) error {
	if err := s.BaseStage.PrepareImage(ctx, c, cr, prevBuiltImage, stageImage); err != nil {
		return err
	}

	if c.GiterminismManager().Dev() {
		addLabels := map[string]string{imagePkg.WerfDevLabel: "true"}
		if c.UseLegacyStapelBuilder(cr) {
			stageImage.Builder.LegacyStapelStageBuilder().BuilderContainer().AddLabel(addLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
		}
	}

	return nil
}
