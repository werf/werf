package container_backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAI_LegacyStageImageContainer_imageRef(t *testing.T) {
	const (
		tmpUUIDName    = "80324e3c-81f8-43cc-a055-6a95a7928690"
		committedID    = "sha256:deadbeef"
		targetPlatform = "linux/amd64"
	)

	t.Run("committed image with target platform returns committed id, not name", func(t *testing.T) {
		img := NewLegacyStageImage(nil, tmpUUIDName, nil, targetPlatform)
		img.buildImage = newLegacyBaseImage(committedID, nil)
		img.builtID = committedID

		assert.Equal(t, committedID, img.container.imageRef(img))
	})

	t.Run("committed image without target platform returns committed id", func(t *testing.T) {
		img := NewLegacyStageImage(nil, tmpUUIDName, nil, "")
		img.buildImage = newLegacyBaseImage(committedID, nil)
		img.builtID = committedID

		assert.Equal(t, committedID, img.container.imageRef(img))
	})

	t.Run("non-committed image with target platform returns name", func(t *testing.T) {
		img := NewLegacyStageImage(nil, "registry.example.com/project:stage", nil, targetPlatform)

		assert.Equal(t, "registry.example.com/project:stage", img.container.imageRef(img))
	})

	t.Run("non-committed image without target platform returns id from name", func(t *testing.T) {
		img := NewLegacyStageImage(nil, "registry.example.com/project:stage", nil, "")

		assert.Equal(t, "registry.example.com/project:stage", img.container.imageRef(img))
	})
}
