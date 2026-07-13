package container_backend

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("LegacyStageImage", func() {
	It("returns stage image name", func() {
		t := GinkgoT()

		stageImage := NewLegacyStageImage("repo:tag", nil, "")
		stageImage.SetStageDesc(&image.StageDesc{
			Info: &image.Info{
				Name: "repo:tag",
				ID:   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		})

		require.NotNil(t, stageImage.GetStageDesc())
		assert.Equal(t, "repo:tag", stageImage.GetID())
	})
})
