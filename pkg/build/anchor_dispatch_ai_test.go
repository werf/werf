package build

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/build/stage"
	imagePkg "github.com/werf/werf/v2/pkg/image"
)

type anchorStub struct {
	*stage.BaseStage
	selectFn func(context.Context, stage.Conveyor, imagePkg.StageDescSet) (*imagePkg.StageDesc, error)
}

func (s *anchorStub) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, set imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	if s.selectFn != nil {
		return s.selectFn(ctx, c, set)
	}
	return s.BaseStage.SelectSuitableStageDesc(ctx, c, set)
}

func TestAI_AnchorSelector_HonoursParentTsToggle(t *testing.T) {
	phase := &BuildPhase{StagesIterator: NewStagesIterator(nil)}

	anchor := &anchorStub{BaseStage: stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{})}
	anchor.SetContentAnchor(true)
	require.Equal(t, int64(0), phase.getPrevNonEmptyStageCreationTsForStage(anchor),
		"anchor stage must resolve with parentStageCreationTs=0")

	nonAnchor := &anchorStub{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{})}
	require.Equal(t, int64(0), phase.getPrevNonEmptyStageCreationTsForStage(nonAnchor),
		"non-anchor branch resolves through the iterator (nil PrevNonEmptyStage → 0)")
}
