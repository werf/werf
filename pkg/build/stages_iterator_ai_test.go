package build

import (
	"testing"

	"github.com/stretchr/testify/require"

	build_image "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
)

// hasPrevStub reports HasPrevStage()==true; used to force the code path that
// touches iterator.PrevNonEmptyStage / iterator.PrevBuiltStage.
type hasPrevStub struct {
	*stage.BaseStage
}

func TestAI_StagesIterator_GetPrevImage_NilPrevIsSafe(t *testing.T) {
	iter := NewStagesIterator(nil)
	stg := &hasPrevStub{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{})}
	require.True(t, stg.HasPrevStage())

	img := &build_image.Image{}
	require.NotPanics(t, func() {
		got := iter.GetPrevImage(img, stg)
		require.Nil(t, got, "pre-iterate anchor resolve gets nil prev image without panicking")
	})
}

func TestAI_StagesIterator_GetPrevBuiltImage_NilPrevIsSafe(t *testing.T) {
	iter := NewStagesIterator(nil)
	stg := &hasPrevStub{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{})}

	img := &build_image.Image{}
	require.NotPanics(t, func() {
		got := iter.GetPrevBuiltImage(img, stg)
		require.Nil(t, got, "pre-iterate anchor resolve gets nil prev built image without panicking")
	})
}
