package image

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/dockerfile"
	"github.com/werf/werf/v2/pkg/dockerfile/frontend"
)

func TestAI_MapDockerfileToImages_RecordsStageRefDependenciesAndDedupes(t *testing.T) {
	dockerfileData := []byte(`
FROM alpine AS base
RUN echo base

FROM base AS builder
RUN echo builder

FROM alpine AS final
COPY --from=builder /b /b
RUN --mount=from=base,target=/mnt echo hi
`)

	d, err := frontend.ParseDockerfileWithBuildkit(util.Sha256Hash("Dockerfile"), dockerfileData, "app", dockerfile.DockerfileOptions{
		Target:         "final",
		TargetPlatform: "linux/amd64",
	})
	require.NoError(t, err)

	images, err := mapDockerfileToImages(context.Background(), d, &config.Meta{}, &config.ImageFromDockerfile{Name: "app", Target: "final"}, "linux/amd64", false, CommonImageOptions{})
	require.NoError(t, err)

	byName := map[string]*Image{}
	for _, img := range images {
		require.NotContains(t, byName, img.Name, "stage %q must be mapped exactly once", img.Name)
		byName[img.Name] = img
	}
	require.Len(t, byName, 3)

	require.ElementsMatch(t, []string{"app/stage/builder", "app/stage/base"}, byName["app"].GetDependencyNames())
	require.ElementsMatch(t, []string{"app/stage/base"}, byName["app/stage/builder"].GetDependencyNames())
	require.Empty(t, byName["app/stage/base"].GetDependencyNames())
}
