package container_backend

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("LegacyStageImage", func() {
	It("prefers stage image ID when available", func() {
		t := GinkgoT()

		stageImage := NewLegacyStageImage(nil, "repo:tag", nil, "")
		stageImage.SetStageDesc(&image.StageDesc{
			Info: &image.Info{
				Name: "repo:tag",
				ID:   "sha256:expected",
			},
		})

		require.NotNil(t, stageImage.GetStageDesc())
		assert.Equal(t, "sha256:expected", stageImage.GetID())
	})

	It("falls back to image name when stage image ID is missing", func() {
		t := GinkgoT()

		stageImage := NewLegacyStageImage(nil, "repo:tag", nil, "")
		stageImage.SetStageDesc(&image.StageDesc{
			Info: &image.Info{
				Name: "repo:tag",
			},
		})

		require.NotNil(t, stageImage.GetStageDesc())
		assert.Equal(t, "repo:tag", stageImage.GetID())
	})
})
