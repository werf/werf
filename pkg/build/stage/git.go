package stage

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_inspector"
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

func (s *GitStage) IsEmpty(ctx context.Context, _ Conveyor, _ container_runtime.ImageInterface) (bool, error) {
	return s.isEmpty(ctx), nil
}

func (s *GitStage) isEmpty(_ context.Context) bool {
	return len(s.gitMappings) == 0
}

func (s *GitStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.BaseStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	if giterminism_inspector.DevMode {
		image.DockerfileImageBuilder().AppendBuildArgs(fmt.Sprintf("--label=%s=true", imagePkg.WerfDevLabel))
	}

	return nil
}
