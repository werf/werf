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

func TestAdoptResolvedBaseImageReference_TargetsWithoutOwnInstructionsStayDistinct(t *testing.T) {
	dockerfileData := []byte(`
FROM alpine AS base

FROM base AS base_with_cmd
CMD echo "from base"

FROM base AS no_cmd
FROM base_with_cmd AS inherited_cmd
`)

	ctx := context.Background()

	contentInputs := func(target, resolvedBaseReference string) []string {
		d, err := frontend.ParseDockerfileWithBuildkit(util.Sha256Hash("Dockerfile"), dockerfileData, target, dockerfile.DockerfileOptions{
			Target:         target,
			TargetPlatform: "linux/amd64",
		})
		require.NoError(t, err)

		images, err := mapDockerfileToImages(ctx, d, &config.Meta{}, &config.ImageFromDockerfile{Name: target, Target: target, Staged: true}, "linux/amd64", false, CommonImageOptions{})
		require.NoError(t, err)

		var targetImage *Image
		for _, img := range images {
			if img.Name == target {
				targetImage = img
			}
		}
		require.NotNil(t, targetImage, "target %q must be mapped to an image", target)
		require.Len(t, targetImage.GetStages(), 1, "target %q must carry no instruction of its own", target)

		targetImage.baseImageReference = resolvedBaseReference
		targetImage.adoptResolvedBaseImageReference()

		var inputs []string
		for _, stg := range targetImage.GetStages() {
			deps, err := stg.GetContentDependencies(ctx, nil, nil)
			require.NoError(t, err)
			inputs = append(inputs, deps)
		}

		return inputs
	}

	require.NotEqual(t,
		contentInputs("no_cmd", "registry.example.com/project:base-content-tag"),
		contentInputs("inherited_cmd", "registry.example.com/project:base-with-cmd-content-tag"),
		"targets sitting on different bases must not share content inputs",
	)
}
