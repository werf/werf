package build

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/container_backend"
	imagePkg "github.com/werf/werf/v3/pkg/image"
)

type slowContentDependenciesStub struct {
	*contentDependenciesStub
	delay time.Duration
}

var _ stage.Interface = (*slowContentDependenciesStub)(nil)

func (s *slowContentDependenciesStub) GetContentDependencies(ctx context.Context, c stage.Conveyor, archive container_backend.BuildContextArchiver) (string, error) {
	time.Sleep(s.delay)

	return s.contentDependenciesStub.GetContentDependencies(ctx, c, archive)
}

// The graph is one dependent pair plus independent images: the dependency is
// the slowest image and comes first, its dependent comes last, so a prepass
// that ignores the graph hands them to different workers and leaves the
// dependent without a digest.
func newAnchorPrepassPhase(parallel bool) (*BuildPhase, []*image.Image) {
	storageManager := &anchorLookupStorageManager{
		primaryStagesStorage:   &anchorPrimaryStagesStorage{},
		secondaryStagesStorage: &fakeStagesStorage{},
		inPrimary:              imagePkg.NewStageDescSet(),
		inSecondary:            imagePkg.NewStageDescSet(),
	}
	phase := newTestBuildPhase(storageManager, nil)
	phase.Conveyor.Parallel = parallel
	phase.Conveyor.ParallelTasksLimit = 4

	var images []*image.Image
	for i := range 6 {
		name := fmt.Sprintf("image%d", i)
		var dependencyNames []string
		if i == 5 {
			dependencyNames = append(dependencyNames, "image0")
		}
		img := newTestImage(name, true, dependencyNames...)
		img.Conveyor = phase.Conveyor
		anchor := &slowContentDependenciesStub{contentDependenciesStub: newContentDependenciesStub(stage.ImageSpec, name)}
		if i == 0 {
			anchor.delay = 300 * time.Millisecond
		}
		anchor.SetContentAnchor(true)
		img.SetStages([]stage.Interface{anchor})
		images = append(images, img)
	}
	phase.Conveyor.imagesTree.SetImagesGraphForTests(newTestImagesGraph(images...))

	return phase, images
}

var _ = ginkgo.Describe("Anchor prepass", func() {
	ginkgo.It("calculates the same digests whether or not the build is parallel", func(ctx ginkgo.SpecContext) {
		serialPhase, serialImages := newAnchorPrepassPhase(false)
		gomega.Expect(serialPhase.calculateAnchorDigests(ctx)).To(gomega.Succeed())

		parallelPhase, parallelImages := newAnchorPrepassPhase(true)
		gomega.Expect(parallelPhase.calculateAnchorDigests(ctx)).To(gomega.Succeed())

		for i, img := range parallelImages {
			gomega.Expect(img.GetAnchorDigest()).NotTo(gomega.BeEmpty(), img.Name)
			gomega.Expect(img.GetAnchorDigest()).To(gomega.Equal(serialImages[i].GetAnchorDigest()), img.Name)
		}
	})

	ginkgo.It("resolves every anchor exactly once in parallel", func(ctx ginkgo.SpecContext) {
		phase, images := newAnchorPrepassPhase(true)
		for i, img := range images {
			img.SetAnchorDigest(fmt.Sprintf("anchor%d", i))
		}
		storageManager := phase.Conveyor.StorageManager.(*anchorLookupStorageManager)

		gomega.Expect(phase.resolveAvailableContentAnchors(ctx)).To(gomega.Succeed())

		gomega.Expect(storageManager.primaryLookups).To(gomega.BeZero())
		gomega.Expect(storageManager.secondaryLookups).To(gomega.BeZero())
		gomega.Expect(storageManager.cachedPrimaryLookups).To(gomega.Equal(len(images)))
		gomega.Expect(storageManager.cachedSecondaryLookups).To(gomega.Equal(len(images)))
	})
})

var _ = ginkgo.Describe("Anchor prepass workers", func() {
	ginkgo.DescribeTable("bounds concurrency by the build parallelism", func(parallel bool, limit int64, tasks, expected int) {
		phase := newTestBuildPhase(nil, nil)
		phase.Conveyor.Parallel = parallel
		phase.Conveyor.ParallelTasksLimit = limit

		gomega.Expect(phase.prepassWorkers(tasks)).To(gomega.Equal(expected))
	},
		ginkgo.Entry("sequential build", false, int64(4), 10, 1),
		ginkgo.Entry("single image", true, int64(4), 1, 1),
		ginkgo.Entry("limit below the task count", true, int64(4), 10, 4),
		ginkgo.Entry("limit above the task count", true, int64(16), 10, 10),
		ginkgo.Entry("unlimited build", true, int64(0), 10, 10),
	)
})
