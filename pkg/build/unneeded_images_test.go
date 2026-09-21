package build

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/config"
)

func newTestImage(t *testing.T, name string, isFinal bool, dependencyNames ...string) *image.Image {
	t.Helper()

	return newTestImageForPlatform(t, "linux/amd64", name, isFinal, dependencyNames...)
}

func newTestImageForPlatform(t *testing.T, targetPlatform, name string, isFinal bool, dependencyNames ...string) *image.Image {
	t.Helper()

	img, err := image.NewImage(context.Background(), targetPlatform, name, image.NoBaseImage, image.ImageOptions{IsFinal: isFinal})
	require.NoError(t, err)
	for _, depName := range dependencyNames {
		img.AddDependencyName(depName)
	}

	return img
}

func TestMarkUnneededImages(t *testing.T) {
	nothingRequested := func(*image.Image) bool { return false }

	t.Run("non-final image is skipped when every dependent is reused by its anchor", func(t *testing.T) {
		base := newTestImage(t, "base", false)
		app := newTestImage(t, "app", true, "base")

		graph, err := image.BuildImagesGraph([]*image.Image{base, app})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{base: false, app: true}, nothingRequested)

		require.True(t, base.Skipped)
		require.False(t, app.Skipped)
	})

	t.Run("non-final image is built when a dependent has to be built", func(t *testing.T) {
		base := newTestImage(t, "base", false)
		app := newTestImage(t, "app", true, "base")
		other := newTestImage(t, "other", true, "base")

		graph, err := image.BuildImagesGraph([]*image.Image{base, app, other})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{base: false, app: true, other: false}, nothingRequested)

		require.False(t, base.Skipped)
	})

	t.Run("non-final image available by its own anchor is not skipped", func(t *testing.T) {
		base := newTestImage(t, "base", false)
		app := newTestImage(t, "app", true, "base")

		graph, err := image.BuildImagesGraph([]*image.Image{base, app})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{base: true, app: true}, nothingRequested)

		require.False(t, base.Skipped)
	})

	t.Run("explicitly requested image is never skipped", func(t *testing.T) {
		base := newTestImage(t, "base", false)
		app := newTestImage(t, "app", true, "base")

		graph, err := image.BuildImagesGraph([]*image.Image{base, app})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{base: false, app: true}, func(img *image.Image) bool {
			return img.Name == "base"
		})

		require.False(t, base.Skipped)
	})

	t.Run("final image is never skipped", func(t *testing.T) {
		base := newTestImage(t, "base", true)
		app := newTestImage(t, "app", true, "base")

		graph, err := image.BuildImagesGraph([]*image.Image{base, app})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{base: false, app: true}, nothingRequested)

		require.False(t, base.Skipped)
	})

	t.Run("skipping propagates down a chain of non-final images", func(t *testing.T) {
		root := newTestImage(t, "root", false)
		middle := newTestImage(t, "middle", false, "root")
		app := newTestImage(t, "app", true, "middle")

		graph, err := image.BuildImagesGraph([]*image.Image{root, middle, app})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{root: false, middle: false, app: true}, nothingRequested)

		require.True(t, middle.Skipped)
		require.True(t, root.Skipped, "a dependency of a skipped image is not needed either")
	})

	t.Run("image is not skipped on one platform only", func(t *testing.T) {
		baseAmd := newTestImageForPlatform(t, "linux/amd64", "base", false)
		appAmd := newTestImageForPlatform(t, "linux/amd64", "app", true, "base")
		baseArm := newTestImageForPlatform(t, "linux/arm64", "base", false)
		appArm := newTestImageForPlatform(t, "linux/arm64", "app", true, "base")

		graph, err := image.BuildImagesGraph([]*image.Image{baseAmd, appAmd, baseArm, appArm})
		require.NoError(t, err)

		markUnneededImages(graph, map[*image.Image]bool{
			baseAmd: false, appAmd: true,
			baseArm: false, appArm: false,
		}, nothingRequested)

		require.False(t, baseAmd.Skipped)
		require.False(t, baseArm.Skipped)
	})
}

func TestIsRequestedImage(t *testing.T) {
	newPhase := func(requestedNames []string, targets ...IntrospectTarget) *BuildPhase {
		return &BuildPhase{
			BuildPhaseOptions: BuildPhaseOptions{
				BuildOptions: BuildOptions{IntrospectOptions: IntrospectOptions{Targets: targets}},
			},
			BasePhase: BasePhase{Conveyor: &Conveyor{imagesTree: image.NewImagesTree(nil, image.ImagesTreeOptions{
				ImagesToProcess: config.ImagesToProcess{ImageNameList: requestedNames},
			})}},
		}
	}

	img := newTestImage(t, "base", false)

	require.False(t, newPhase(nil).isRequestedImage(img))
	require.True(t, newPhase([]string{"base"}).isRequestedImage(img),
		"an image named on the command line has to be built")
	require.False(t, newPhase([]string{"other"}).isRequestedImage(img))
	require.True(t, newPhase(nil, IntrospectTarget{ImageName: "base", StageName: "install"}).isRequestedImage(img),
		"an image whose stage is introspected has to be built")
	require.True(t, newPhase(nil, IntrospectTarget{ImageName: "*", StageName: "install"}).isRequestedImage(img))
	require.False(t, newPhase(nil, IntrospectTarget{ImageName: "other", StageName: "install"}).isRequestedImage(img))
}

func TestMarkUnneededImages_ChainWithPerPlatformAnchors(t *testing.T) {
	nothingRequested := func(*image.Image) bool { return false }

	// The anchor of the middle image is present for one platform only, so it has
	// to be built for the other one — together with the image it is built from.
	baseAmd := newTestImageForPlatform(t, "linux/amd64", "base", false)
	baseArm := newTestImageForPlatform(t, "linux/arm64", "base", false)
	middleAmd := newTestImageForPlatform(t, "linux/amd64", "middle", false, "base")
	middleArm := newTestImageForPlatform(t, "linux/arm64", "middle", false, "base")
	appAmd := newTestImageForPlatform(t, "linux/amd64", "app", true, "middle")
	appArm := newTestImageForPlatform(t, "linux/arm64", "app", true, "middle")

	graph, err := image.BuildImagesGraph([]*image.Image{baseAmd, middleAmd, appAmd, baseArm, middleArm, appArm})
	require.NoError(t, err)

	markUnneededImages(graph, map[*image.Image]bool{
		baseAmd: false, baseArm: false,
		middleAmd: false, middleArm: true,
		appAmd: true, appArm: true,
	}, nothingRequested)

	require.False(t, middleAmd.Skipped, "the anchor of middle is absent for this platform, so it is built")
	require.False(t, middleArm.Skipped, "an image name is skipped for all of its platforms or for none")
	require.False(t, baseAmd.Skipped, "middle is built on this platform and needs base")
	require.False(t, baseArm.Skipped)
}
