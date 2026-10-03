package build

import (
	"context"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/giterminism_manager"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/manager"
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

var _ giterminism_manager.Interface = (*checkModeGiterminismManager)(nil)

type checkModeGiterminismManager struct {
	giterminism_manager.Interface
}

func (m *checkModeGiterminismManager) HeadCommit(_ context.Context) string { return "headcommit" }

var _ manager.StorageManagerInterface = (*checkModeStorageManager)(nil)

type checkModeStorageManager struct {
	manager.StorageManagerInterface
	stagesStorage *checkModeStorage
}

func (m *checkModeStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return m.stagesStorage
}

func (m *checkModeStorageManager) GetMetaStorage() storage.PrimaryStagesStorage {
	return m.stagesStorage
}

func (m *checkModeStorageManager) GetFinalStagesStorage() storage.StagesStorage { return nil }

var _ storage.PrimaryStagesStorage = (*checkModeStorage)(nil)

// checkModeStorage records every mutating call so that a check-mode spec can assert that none
// happened, instead of asserting that one particular write is skipped.
type checkModeStorage struct {
	storage.PrimaryStagesStorage
	desc   *imagePkg.StageDesc
	writes []string
}

func (s *checkModeStorage) String() string { return "check-mode-test" }

func (s *checkModeStorage) ConstructStageImageName(projectName, digest string, creationTs int64) string {
	return projectName + ":" + digest
}

func (s *checkModeStorage) GetStageDesc(_ context.Context, _ string, _ imagePkg.StageID) (*imagePkg.StageDesc, error) {
	if s.desc == nil {
		return nil, storage.ErrStageNotFound
	}
	return s.desc, nil
}

func (s *checkModeStorage) IsManagedImageExist(_ context.Context, _, _ string, _ ...storage.Option) (bool, error) {
	return false, nil
}

func (s *checkModeStorage) IsImageMetadataExist(_ context.Context, _, _, _, _ string, _ ...storage.Option) (bool, error) {
	return false, nil
}

func (s *checkModeStorage) CheckStageCustomTag(_ context.Context, _ *imagePkg.StageDesc, _ string) error {
	return nil
}

func (s *checkModeStorage) AddManagedImage(_ context.Context, _, _ string) error {
	s.writes = append(s.writes, "AddManagedImage")
	return nil
}

func (s *checkModeStorage) PutImageMetadata(_ context.Context, _, _, _, _ string) error {
	s.writes = append(s.writes, "PutImageMetadata")
	return nil
}

func (s *checkModeStorage) PostMultiplatformImage(_ context.Context, _, _ string, _ []*imagePkg.Info, _ []string) error {
	s.writes = append(s.writes, "PostMultiplatformImage")
	return nil
}

func (s *checkModeStorage) AddStageCustomTag(_ context.Context, _ *imagePkg.StageDesc, _ string) error {
	s.writes = append(s.writes, "AddStageCustomTag")
	return nil
}

func (s *checkModeStorage) RegisterStageCustomTag(_ context.Context, _ string, _ *imagePkg.StageDesc, _ string) error {
	s.writes = append(s.writes, "RegisterStageCustomTag")
	return nil
}
