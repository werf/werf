package build

import (
	"context"
	"errors"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/config"
	imagePkg "github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.Describe("Check mode image publication", func() {
	stageDesc := &imagePkg.StageDesc{
		StageID: imagePkg.NewStageID("digest0", 1),
		Info: &imagePkg.Info{
			Name:       "repo:digest0-1",
			Tag:        "digest0-1",
			Repository: "repo",
			Labels:     map[string]string{imagePkg.WerfStageContentDigestLabel: "content"},
		},
	}

	newPhase := func(shouldBeBuiltMode bool) (*BuildPhase, *checkModeStorage) {
		stagesStorage := &checkModeStorage{desc: stageDesc}
		phase := NewBuildPhase(&Conveyor{
			StorageManager:     &checkModeStorageManager{stagesStorage: stagesStorage},
			werfConfig:         &config.WerfConfig{Meta: &config.Meta{Project: "check-mode-project"}},
			giterminismManager: &checkModeGiterminismManager{},
			serviceRWMutex:     make(map[string]*sync.RWMutex),
			stageDigestMutex:   make(map[string]*sync.Mutex),
		}, BuildPhaseOptions{
			ShouldBeBuiltMode: shouldBeBuiltMode,
			BuildOptions: BuildOptions{
				CustomTagFuncList: []imagePkg.CustomTagFunc{func(_, _ string) string { return "custom" }},
			},
		})
		phase.Conveyor.SetShouldAddManagedImagesRecords()

		return phase, stagesStorage
	}

	singlePlatformImage := func() *image.Image {
		img, err := image.NewImage(context.Background(), "linux/amd64", "image0", image.NoBaseImage, image.ImageOptions{IsFinal: true, UseCustomTag: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		img.SetContentTagDesc(stageDesc)

		return img
	}

	multiplatformImage := func() *image.MultiplatformImage {
		return image.NewMultiplatformImage("image0", []*image.Image{singlePlatformImage()}, 0, 1)
	}

	ginkgo.It("writes nothing for a single-platform image", func(ctx ginkgo.SpecContext) {
		phase, stagesStorage := newPhase(true)
		gomega.Expect(phase.publishImageMetadata(ctx, "image0", singlePlatformImage())).To(gomega.Succeed())
		gomega.Expect(stagesStorage.writes).To(gomega.BeEmpty())
	})

	ginkgo.It("writes nothing for a multiplatform image", func(ctx ginkgo.SpecContext) {
		phase, stagesStorage := newPhase(true)
		img := multiplatformImage()
		gomega.Expect(phase.publishMultiplatformImageMetadata(ctx, "image0", img)).To(gomega.Succeed())
		gomega.Expect(phase.publishMultiplatformImageCustomTags(ctx, "image0", img)).To(gomega.Succeed())
		gomega.Expect(stagesStorage.writes).To(gomega.BeEmpty())
	})

	ginkgo.DescribeTable("rejects a missing custom tag in check mode", func(ctx ginkgo.SpecContext, multiplatform bool) {
		phase, stagesStorage := newPhase(true)
		stagesStorage.customTagErr = errors.New("custom tag missing")
		var err error
		if multiplatform {
			img := multiplatformImage()
			gomega.Expect(phase.publishMultiplatformImageMetadata(ctx, "image0", img)).To(gomega.Succeed())
			err = phase.publishMultiplatformImageCustomTags(ctx, "image0", img)
		} else {
			err = phase.publishImageMetadata(ctx, "image0", singlePlatformImage())
		}
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("custom tag missing")))
		gomega.Expect(stagesStorage.writes).To(gomega.BeEmpty())
	},
		ginkgo.Entry("single-platform image", false),
		ginkgo.Entry("multiplatform image", true),
	)

	ginkgo.It("reports a multiplatform image that was never published", func(ctx ginkgo.SpecContext) {
		phase, stagesStorage := newPhase(true)
		stagesStorage.desc = nil
		err := phase.publishMultiplatformImageMetadata(ctx, "image0", multiplatformImage())
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("is not published")))
		gomega.Expect(stagesStorage.writes).To(gomega.BeEmpty())
	})

	ginkgo.DescribeTable("publishes outside check mode", func(ctx ginkgo.SpecContext, publish func(phase *BuildPhase) error, expectedWrites []string) {
		phase, stagesStorage := newPhase(false)
		gomega.Expect(publish(phase)).To(gomega.Succeed())
		gomega.Expect(stagesStorage.writes).To(gomega.ConsistOf(expectedWrites))
	},
		ginkgo.Entry("a single-platform image", func(phase *BuildPhase) error {
			return phase.publishImageMetadata(context.Background(), "image0", singlePlatformImage())
		}, []string{"AddManagedImage", "PutImageMetadata", "AddStageCustomTag", "RegisterStageCustomTag"}),
		ginkgo.Entry("a multiplatform image", func(phase *BuildPhase) error {
			img := multiplatformImage()
			if err := phase.publishMultiplatformImageMetadata(context.Background(), "image0", img); err != nil {
				return err
			}
			return phase.publishMultiplatformImageCustomTags(context.Background(), "image0", img)
		}, []string{"AddManagedImage", "PostMultiplatformImage", "PutImageMetadata", "AddStageCustomTag", "RegisterStageCustomTag"}),
	)
})
