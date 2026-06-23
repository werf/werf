package image

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	common_image "github.com/werf/werf/v2/pkg/image"
)

func newTestImage(t *testing.T, targetPlatform, name string) *Image {
	t.Helper()

	img, err := NewImage(context.Background(), targetPlatform, name, NoBaseImage, ImageOptions{})
	require.NoError(t, err)
	return img
}

func TestAI_NewMultiplatformImage_PanicsWhenContentTagDescMissing(t *testing.T) {
	img := newTestImage(t, "linux/amd64", "app")

	require.PanicsWithValue(
		t,
		`content tag descriptor is not set for image "app" platform "linux/amd64"`,
		func() { NewMultiplatformImage("app", []*Image{img}, 0, 1) },
	)
}

func TestAI_NewMultiplatformImage_DeterministicDigestRegardlessOfPlatformOrder(t *testing.T) {
	amd := newTestImage(t, "linux/amd64", "app")
	amd.SetContentTagDesc(&common_image.StageDesc{
		StageID: common_image.NewStageID("digest-amd", 1),
		Info:    &common_image.Info{},
	})

	arm := newTestImage(t, "linux/arm64", "app")
	arm.SetContentTagDesc(&common_image.StageDesc{
		StageID: common_image.NewStageID("digest-arm", 1),
		Info:    &common_image.Info{},
	})

	first := NewMultiplatformImage("app", []*Image{amd, arm}, 0, 1)
	second := NewMultiplatformImage("app", []*Image{arm, amd}, 0, 1)

	require.Equal(t, first.GetDigest(), second.GetDigest())
}
