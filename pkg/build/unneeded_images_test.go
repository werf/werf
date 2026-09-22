package build

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

func newTestImage(name string, isFinal bool, dependencyNames ...string) *image.Image {
	return newTestImageForPlatform("linux/amd64", name, isFinal, dependencyNames...)
}

func newTestImageForPlatform(targetPlatform, name string, isFinal bool, dependencyNames ...string) *image.Image {
	img, err := image.NewImage(context.Background(), targetPlatform, name, image.NoBaseImage, image.ImageOptions{IsFinal: isFinal})
	Expect(err).NotTo(HaveOccurred())
	for _, depName := range dependencyNames {
		img.AddDependencyName(depName)
	}

	return img
}

func newTestImagesGraph(images ...*image.Image) *image.ImagesGraph {
	graph, err := image.BuildImagesGraph(images)
	Expect(err).NotTo(HaveOccurred())

	return graph
}

func newTestBuildPhase(storageManager manager.StorageManagerInterface, requestedNames []string, introspectTargets ...IntrospectTarget) *BuildPhase {
	return &BuildPhase{
		BuildPhaseOptions: BuildPhaseOptions{
			BuildOptions: BuildOptions{IntrospectOptions: IntrospectOptions{Targets: introspectTargets}},
		},
		BasePhase: BasePhase{Conveyor: &Conveyor{
			StorageManager: storageManager,
			imagesTree: image.NewImagesTree(nil, image.ImagesTreeOptions{
				ImagesToProcess: config.ImagesToProcess{ImageNameList: requestedNames},
			}),
		}},
	}
}

type unneededImagesScenario struct {
	graph         *image.ImagesGraph
	anchorExists  map[*image.Image]bool
	isRequested   func(img *image.Image) bool
	expectSkipped map[*image.Image]bool
}

func nothingRequested(*image.Image) bool { return false }

var _ = Describe("markUnneededImages", func() {
	DescribeTable("deciding which images no image being built needs",
		func(setup func() unneededImagesScenario) {
			scenario := setup()

			markUnneededImages(scenario.graph, scenario.anchorExists, scenario.isRequested)

			for img, expected := range scenario.expectSkipped {
				Expect(img.Skipped).To(Equal(expected), fmt.Sprintf("image %s (%s)", img.Name, img.TargetPlatform))
			}
		},
		Entry("non-final image is skipped when every dependent is reused by its anchor", func() unneededImagesScenario {
			base := newTestImage("base", false)
			app := newTestImage("app", true, "base")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(base, app),
				anchorExists:  map[*image.Image]bool{base: false, app: true},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{base: true, app: false},
			}
		}),
		Entry("non-final image is built when a dependent has to be built", func() unneededImagesScenario {
			base := newTestImage("base", false)
			app := newTestImage("app", true, "base")
			other := newTestImage("other", true, "base")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(base, app, other),
				anchorExists:  map[*image.Image]bool{base: false, app: true, other: false},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{base: false},
			}
		}),
		Entry("non-final image available by its own anchor is not skipped", func() unneededImagesScenario {
			base := newTestImage("base", false)
			app := newTestImage("app", true, "base")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(base, app),
				anchorExists:  map[*image.Image]bool{base: true, app: true},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{base: false},
			}
		}),
		Entry("explicitly requested image is never skipped", func() unneededImagesScenario {
			base := newTestImage("base", false)
			app := newTestImage("app", true, "base")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(base, app),
				anchorExists:  map[*image.Image]bool{base: false, app: true},
				isRequested:   func(img *image.Image) bool { return img.Name == "base" },
				expectSkipped: map[*image.Image]bool{base: false},
			}
		}),
		Entry("final image is never skipped", func() unneededImagesScenario {
			base := newTestImage("base", true)
			app := newTestImage("app", true, "base")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(base, app),
				anchorExists:  map[*image.Image]bool{base: false, app: true},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{base: false},
			}
		}),
		Entry("skipping propagates down a chain of non-final images", func() unneededImagesScenario {
			root := newTestImage("root", false)
			middle := newTestImage("middle", false, "root")
			app := newTestImage("app", true, "middle")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(root, middle, app),
				anchorExists:  map[*image.Image]bool{root: false, middle: false, app: true},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{middle: true, root: true},
			}
		}),
		Entry("nothing is skipped in a chain whose anchors are all absent", func() unneededImagesScenario {
			base := newTestImage("base", false)
			middle := newTestImage("middle", false, "base")
			app := newTestImage("app", true, "middle")

			return unneededImagesScenario{
				graph:         newTestImagesGraph(base, middle, app),
				anchorExists:  map[*image.Image]bool{base: false, middle: false, app: false},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{base: false, middle: false, app: false},
			}
		}),
		Entry("image is not skipped on one platform only", func() unneededImagesScenario {
			baseAmd := newTestImageForPlatform("linux/amd64", "base", false)
			appAmd := newTestImageForPlatform("linux/amd64", "app", true, "base")
			baseArm := newTestImageForPlatform("linux/arm64", "base", false)
			appArm := newTestImageForPlatform("linux/arm64", "app", true, "base")

			return unneededImagesScenario{
				graph: newTestImagesGraph(baseAmd, appAmd, baseArm, appArm),
				anchorExists: map[*image.Image]bool{
					baseAmd: false, appAmd: true,
					baseArm: false, appArm: false,
				},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{baseAmd: false, baseArm: false},
			}
		}),
		Entry("chain with per-platform anchors", func() unneededImagesScenario {
			baseAmd := newTestImageForPlatform("linux/amd64", "base", false)
			baseArm := newTestImageForPlatform("linux/arm64", "base", false)
			middleAmd := newTestImageForPlatform("linux/amd64", "middle", false, "base")
			middleArm := newTestImageForPlatform("linux/arm64", "middle", false, "base")
			appAmd := newTestImageForPlatform("linux/amd64", "app", true, "middle")
			appArm := newTestImageForPlatform("linux/arm64", "app", true, "middle")

			return unneededImagesScenario{
				graph: newTestImagesGraph(baseAmd, middleAmd, appAmd, baseArm, middleArm, appArm),
				anchorExists: map[*image.Image]bool{
					baseAmd: false, baseArm: false,
					middleAmd: false, middleArm: true,
					appAmd: true, appArm: true,
				},
				isRequested:   nothingRequested,
				expectSkipped: map[*image.Image]bool{middleAmd: false, middleArm: false, baseAmd: false, baseArm: false},
			}
		}),
	)

	It("skips nothing when every image name is requested, as with --final-images-only=false", func() {
		base := newTestImage("base", false)
		app := newTestImage("app", true, "base")
		phase := newTestBuildPhase(nil, []string{"base", "app"})

		markUnneededImages(newTestImagesGraph(base, app), map[*image.Image]bool{base: false, app: true}, phase.isRequestedImage)

		Expect(base.Skipped).To(BeFalse())
		Expect(app.Skipped).To(BeFalse())
	})

	It("skips a missing internal staged subimage when the requested configured image is reused", func() {
		internal := newTestImage("app/stage/build", false)
		app := newTestImage("app", true, "app/stage/build")
		phase := newTestBuildPhase(nil, []string{"app"})

		markUnneededImages(newTestImagesGraph(internal, app), map[*image.Image]bool{internal: false, app: true}, phase.isRequestedImage)

		Expect(internal.Skipped).To(BeTrue())
		Expect(app.Skipped).To(BeFalse())
	})
})

var _ = Describe("BuildPhase.isRequestedImage", func() {
	var img *image.Image

	BeforeEach(func() {
		img = newTestImage("base", false)
	})

	It("is false when nothing is requested", func() {
		Expect(newTestBuildPhase(nil, nil).isRequestedImage(img)).To(BeFalse())
	})

	It("is true for an image named on the command line", func() {
		Expect(newTestBuildPhase(nil, []string{"base"}).isRequestedImage(img)).To(BeTrue())
	})

	It("is false for another image named on the command line", func() {
		Expect(newTestBuildPhase(nil, []string{"other"}).isRequestedImage(img)).To(BeFalse())
	})

	It("is true for an image whose stage is introspected", func() {
		Expect(newTestBuildPhase(nil, nil, IntrospectTarget{ImageName: "base", StageName: "install"}).isRequestedImage(img)).To(BeTrue())
	})

	It("is true when every image is introspected", func() {
		Expect(newTestBuildPhase(nil, nil, IntrospectTarget{ImageName: "*", StageName: "install"}).isRequestedImage(img)).To(BeTrue())
	})

	It("is false when another image is introspected", func() {
		Expect(newTestBuildPhase(nil, nil, IntrospectTarget{ImageName: "other", StageName: "install"}).isRequestedImage(img)).To(BeFalse())
	})
})

var _ manager.StorageManagerInterface = (*anchorLookupStorageManager)(nil)

type anchorLookupStorageManager struct {
	manager.StorageManagerInterface
	secondaryStagesStorage storage.StagesStorage
	inPrimary              imagePkg.StageDescSet
	inSecondary            imagePkg.StageDescSet
}

func (m *anchorLookupStorageManager) GetStageDescSetByDigestWithCache(_ context.Context, _, _ string, _ int64) (imagePkg.StageDescSet, error) {
	return m.inPrimary, nil
}

func (m *anchorLookupStorageManager) GetSecondaryStagesStorageList() []storage.StagesStorage {
	return []storage.StagesStorage{m.secondaryStagesStorage}
}

func (m *anchorLookupStorageManager) GetStageDescSetByDigestFromStagesStorageWithCache(_ context.Context, _, _ string, _ int64, stagesStorage storage.StagesStorage) (imagePkg.StageDescSet, error) {
	if stagesStorage == m.secondaryStagesStorage {
		return m.inSecondary, nil
	}

	return m.inPrimary, nil
}

func (m *anchorLookupStorageManager) SelectSuitableStageDesc(_ context.Context, _ stage.Conveyor, _ stage.Interface, stageDescSet imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	for stageDesc := range stageDescSet.Iter() {
		return stageDesc, nil
	}

	return nil, nil
}

var _ = Describe("BuildPhase.anchorExistsInStagesStorage", func() {
	newAnchoredImage := func() *image.Image {
		img := newTestImage("base", false)
		img.SetAnchorDigest("anchor-digest")
		img.SetStages([]stage.Interface{stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{ImageName: "base"})})

		return img
	}

	anchorStageDesc := &imagePkg.StageDesc{
		StageID: imagePkg.NewStageID("anchor-digest", 1),
		Info:    &imagePkg.Info{Name: "repo:anchor"},
	}

	DescribeTable("looking the content anchor up",
		func(inPrimary, inSecondary imagePkg.StageDescSet, expected bool) {
			phase := newTestBuildPhase(&anchorLookupStorageManager{
				secondaryStagesStorage: &fakeStagesStorage{},
				inPrimary:              inPrimary,
				inSecondary:            inSecondary,
			}, nil)

			exists, err := phase.anchorExistsInStagesStorage(context.Background(), newAnchoredImage())

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(Equal(expected))
		},
		Entry("found in the primary stages storage", imagePkg.NewStageDescSet(anchorStageDesc), imagePkg.NewStageDescSet(), true),
		Entry("found in a secondary stages storage only", imagePkg.NewStageDescSet(), imagePkg.NewStageDescSet(anchorStageDesc), true),
		Entry("found nowhere", imagePkg.NewStageDescSet(), imagePkg.NewStageDescSet(), false),
	)

	It("is false for an image without a content anchor digest", func() {
		img := newAnchoredImage()
		img.SetAnchorDigest("")

		exists, err := newTestBuildPhase(nil, nil).anchorExistsInStagesStorage(context.Background(), img)

		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
	})
})

var _ storage.StagesStorage = (*fakeStagesStorage)(nil)

type fakeStagesStorage struct {
	storage.StagesStorage
}
