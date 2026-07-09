package build

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/container_backend"
)

func TestAI_ValidateBackendStapelSupport(t *testing.T) {
	dockerBackend := container_backend.NewDockerServerBackend(nil)

	t.Run("rejects stapel image on backend without stapel support", func(t *testing.T) {
		err := validateBackendStapelSupport(dockerBackend, []*image.Image{
			{Name: "backend", IsDockerfileImage: false},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), `building of stapel image "backend" is not supported`)
		require.Contains(t, err.Error(), "WERF_BUILDKIT_HOST")
	})

	t.Run("allows dockerfile-only selection from mixed config", func(t *testing.T) {
		err := validateBackendStapelSupport(dockerBackend, []*image.Image{
			{Name: "frontend", IsDockerfileImage: true},
		})
		require.NoError(t, err)
	})
}
