package stage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAI_ContentAnchor_DefaultsFalse(t *testing.T) {
	s := NewBaseStage(From, &BaseStageOptions{})
	require.False(t, s.IsContentAnchor(), "IsContentAnchor default must be false")
}

func TestAI_ContentAnchor_SetterToggles(t *testing.T) {
	s := NewBaseStage(From, &BaseStageOptions{})

	s.SetContentAnchor(true)
	require.True(t, s.IsContentAnchor())

	s.SetContentAnchor(false)
	require.False(t, s.IsContentAnchor())
}

func TestAI_ContentAnchor_PropagatesThroughEmbedding(t *testing.T) {
	// Concrete stage types embed *BaseStage; anchor bit must be visible through
	// the Interface method set the image mapper uses to set it positionally.
	s := &ImageSpecStage{BaseStage: NewBaseStage(ImageSpec, &BaseStageOptions{})}

	var i Interface = s
	require.False(t, i.IsContentAnchor())

	i.SetContentAnchor(true)
	require.True(t, i.IsContentAnchor())
	require.True(t, s.IsContentAnchor(), "concrete stage type sees the same anchor bit")
}
