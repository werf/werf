package build

import (
	"testing"

	"github.com/stretchr/testify/require"

	imagePkg "github.com/werf/werf/v2/pkg/image"
)

func TestAI_ComposeEmptyAnchorLabels_OverridesTakePrecedence(t *testing.T) {
	prev := map[string]string{
		"custom":                            "keep",
		imagePkg.WerfLabel:                  "project",
		imagePkg.WerfParentStageID:          "prev-parent-id",
		imagePkg.WerfStageContentDigestLabel: "prev-content-sig",
	}

	labels := composeEmptyAnchorLabels(prev, "new-parent-id", "anchor-content-sig")

	require.Equal(t, "keep", labels["custom"], "unrelated labels are copied through")
	require.Equal(t, "project", labels[imagePkg.WerfLabel], "werf label is copied through")
	require.Equal(t, "new-parent-id", labels[imagePkg.WerfParentStageID],
		"WerfParentStageID must be overwritten with the caller-supplied value")
	require.Equal(t, "anchor-content-sig", labels[imagePkg.WerfStageContentDigestLabel],
		"WerfStageContentDigestLabel must be the anchor's own content signature")
}

func TestAI_ComposeEmptyAnchorLabels_TolerateNilPrevLabels(t *testing.T) {
	labels := composeEmptyAnchorLabels(nil, "pid", "csig")

	require.Equal(t, map[string]string{
		imagePkg.WerfParentStageID:          "pid",
		imagePkg.WerfStageContentDigestLabel: "csig",
	}, labels)
}

func TestAI_ComposeEmptyAnchorLabels_PanicOnEmptyParent(t *testing.T) {
	require.PanicsWithValue(t, "composeEmptyAnchorLabels: parentStageID must not be empty", func() {
		composeEmptyAnchorLabels(nil, "", "csig")
	})
}

func TestAI_ComposeEmptyAnchorLabels_PanicOnEmptyContentDigest(t *testing.T) {
	require.PanicsWithValue(t, "composeEmptyAnchorLabels: contentDigest must not be empty", func() {
		composeEmptyAnchorLabels(nil, "pid", "")
	})
}
