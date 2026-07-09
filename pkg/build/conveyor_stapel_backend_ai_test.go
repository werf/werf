package build

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
)

func TestAI_ConveyorBuild_RejectsStapelImagesWithoutStapelBuildSupport(t *testing.T) {
	werfConfig := config.NewWerfConfig(&config.Meta{}, []config.ImageInterface{
		&config.StapelImage{StapelImageBase: &config.StapelImageBase{Name: "backend"}},
	})

	c := &Conveyor{
		werfConfig:       werfConfig,
		ContainerBackend: container_backend.NewDockerServerBackend(nil),
	}

	_, err := c.Build(context.Background(), BuildOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), `building of stapel image "backend" is not supported`)
	require.Contains(t, err.Error(), "WERF_BUILDKIT_HOST")
}
