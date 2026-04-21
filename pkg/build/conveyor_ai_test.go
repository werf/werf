package build

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
)

var _ = Describe("Conveyor stage image cache", func() {
	var conveyor *Conveyor

	BeforeEach(func() {
		conveyor = &Conveyor{
			stageImages:      make(map[string]*stage.StageImage),
			serviceRWMutex:   map[string]*sync.RWMutex{},
			stageDigestMutex: map[string]*sync.Mutex{},
		}
	})

	Describe("GetStageImageByPlatform", func() {
		It("returns the correct image per platform", func() {
			amd64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/amd64"))
			arm64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/arm64"))

			conveyor.SetStageImage(amd64Image)
			conveyor.SetStageImage(arm64Image)

			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64")).To(BeIdenticalTo(amd64Image))
			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64")).To(BeIdenticalTo(arm64Image))
			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64")).NotTo(BeIdenticalTo(arm64Image))
		})
	})

	Describe("getStageImage", func() {
		It("returns nil when multiple platforms are cached", func() {
			conveyor.SetStageImage(stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/amd64")))
			conveyor.SetStageImage(stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/arm64")))

			Expect(conveyor.getStageImage("alpine:3.22.3")).To(BeNil())
		})

		It("returns the image when only one platform is cached", func() {
			arm64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/arm64"))
			conveyor.SetStageImage(arm64Image)

			Expect(conveyor.getStageImage("alpine:3.22.3")).To(BeIdenticalTo(arm64Image))
		})
	})

	Describe("UnsetStageImageByPlatform", func() {
		It("removes only the specified platform, leaving others intact", func() {
			amd64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/amd64"))
			arm64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/arm64"))

			conveyor.SetStageImage(amd64Image)
			conveyor.SetStageImage(arm64Image)

			conveyor.UnsetStageImageByPlatform("alpine:3.22.3", "linux/amd64")

			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64")).To(BeNil())
			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64")).To(BeIdenticalTo(arm64Image))
		})
	})

	Describe("UnsetStageImage", func() {
		It("removes all platform variants", func() {
			conveyor.SetStageImage(stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/amd64")))
			conveyor.SetStageImage(stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/arm64")))

			conveyor.UnsetStageImage("alpine:3.22.3")

			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64")).To(BeNil())
			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64")).To(BeNil())
		})
	})

	Describe("GetOrCreateStageImage", func() {
		It("creates separate stage image objects for each platform", func() {
			amd64Img := &image.Image{TargetPlatform: "linux/amd64"}
			arm64Img := &image.Image{TargetPlatform: "linux/arm64"}

			amd64StageImage := conveyor.GetOrCreateStageImage("alpine:3.22.3", nil, nil, amd64Img)
			arm64StageImage := conveyor.GetOrCreateStageImage("alpine:3.22.3", nil, nil, arm64Img)

			Expect(amd64StageImage).NotTo(BeIdenticalTo(arm64StageImage))
			Expect(amd64StageImage.Image.GetTargetPlatform()).To(Equal("linux/amd64"))
			Expect(arm64StageImage.Image.GetTargetPlatform()).To(Equal("linux/arm64"))
		})

		It("returns the cached image on repeated calls for the same platform", func() {
			img := &image.Image{TargetPlatform: "linux/amd64"}

			first := conveyor.GetOrCreateStageImage("alpine:3.22.3", nil, nil, img)
			second := conveyor.GetOrCreateStageImage("alpine:3.22.3", nil, nil, img)

			Expect(first).To(BeIdenticalTo(second))
		})

		// Regression test for the original bug: in a multiplatform build both platforms
		// share the same base image name. After the amd64 stage is built, build_phase.go
		// renames it via UnsetStageImageByPlatform(oldName, platform) + SetStageImage(newName).
		// The old, broken code used UnsetStageImage(oldName) which evicted ALL platforms,
		// so the arm64 stage image was silently lost and the next GetOrCreateStageImage call
		// for arm64 created a brand-new object — breaking the arm64 build pipeline.
		It("arm64 stage survives the amd64 stage rename that happens in build_phase", func() {
			amd64Img := &image.Image{TargetPlatform: "linux/amd64"}
			arm64Img := &image.Image{TargetPlatform: "linux/arm64"}

			amd64StageImage := conveyor.GetOrCreateStageImage("alpine:3.22.3", nil, nil, amd64Img)
			arm64StageImage := conveyor.GetOrCreateStageImage("alpine:3.22.3", nil, nil, arm64Img)

			// Simulate build_phase.go: amd64 stage built, rename to final digest name.
			conveyor.UnsetStageImageByPlatform(amd64StageImage.Image.Name(), "linux/amd64")
			amd64StageImage.Image.SetName("werf-stages-storage/abc123:linux-amd64")
			conveyor.SetStageImage(amd64StageImage)

			// arm64 stage must still be retrievable under the original name.
			Expect(conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64")).To(BeIdenticalTo(arm64StageImage))
		})
	})
})
