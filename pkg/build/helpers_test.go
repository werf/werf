package build

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/lockgate/pkg/distributed_locker/optimistic_locking_store"
	buildImage "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/stage_builder"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
)

func newReportPhase(reportPath string) *BuildPhase {
	return NewBuildPhase(nil, BuildPhaseOptions{BuildOptions: BuildOptions{ReportPath: reportPath, ReportFormat: ReportJSON}})
}

type operationsReport struct {
	Operations      map[string]ReportOperationRecord
	CacheOperations map[string]map[string]ReportCacheOperationRecord
	StageCache      map[string]int
	RegistryCache   map[string]int
	Recovery        map[string]int
}

func decodeOperationsReport(data []byte) operationsReport {
	var decoded operationsReport
	gomega.Expect(json.Unmarshal(data, &decoded)).To(gomega.Succeed())
	return decoded
}

func readOperationsReport(path string) operationsReport {
	data, err := os.ReadFile(path)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return decodeOperationsReport(data)
}

type publicationStorage struct {
	storage.PrimaryStagesStorage
	desc *imagePkg.StageDesc
}

var _ storage.PrimaryStagesStorage = (*publicationStorage)(nil)

func (s *publicationStorage) String() string { return "publication-test" }
func (s *publicationStorage) StoreImage(_ context.Context, img container_backend.LegacyImageInterface) error {
	s.desc = &imagePkg.StageDesc{StageID: imagePkg.NewStageID("shared-digest", 100), Info: &imagePkg.Info{Name: img.Name(), Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "winner-content"}}}
	return nil
}

func (s *publicationStorage) GetStageDesc(_ context.Context, _ string, _ imagePkg.StageID) (*imagePkg.StageDesc, error) {
	return s.desc, nil
}

type publicationStorageManager struct {
	manager.StorageManagerInterface
	primary *publicationStorage
}

var _ manager.StorageManagerInterface = (*publicationStorageManager)(nil)

func (m *publicationStorageManager) GetStagesStorage() storage.PrimaryStagesStorage { return m.primary }
func (m *publicationStorageManager) GetStageDescSetByDigest(_ context.Context, _, _ string, _ int64) (imagePkg.StageDescSet, error) {
	if m.primary.desc == nil {
		return imagePkg.NewStageDescSet(), nil
	}
	return imagePkg.NewStageDescSet(m.primary.desc), nil
}

func (m *publicationStorageManager) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, stg stage.Interface, set imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	return stg.SelectSuitableStageDesc(ctx, c, set)
}

func (m *publicationStorageManager) GenerateStageDescCreationTs(_ string, _ imagePkg.StageDescSet) (string, int64) {
	return "repo:shared-digest-100", 100
}
func (m *publicationStorageManager) GetCacheStagesStorageList() []storage.StagesStorage { return nil }
func (m *publicationStorageManager) CopyStageIntoCacheStorages(_ context.Context, _ imagePkg.StageID, _ []storage.StagesStorage, _ manager.CopyStageIntoStorageOptions) error {
	return nil
}

type publicationImage struct {
	container_backend.LegacyImageInterface
	name      string
	stageDesc *imagePkg.StageDesc
}

var _ container_backend.LegacyImageInterface = (*publicationImage)(nil)

func (i *publicationImage) Name() string                          { return i.name }
func (i *publicationImage) GetTargetPlatform() string             { return "linux/amd64" }
func (i *publicationImage) SetName(name string)                   { i.name = name }
func (i *publicationImage) SetStageDesc(desc *imagePkg.StageDesc) { i.stageDesc = desc }
func (i *publicationImage) GetStageDesc() *imagePkg.StageDesc     { return i.stageDesc }

type publicationStage struct{ *stage.BaseStage }

var _ stage.Interface = (*publicationStage)(nil)

func (s *publicationStage) IsBuildable() bool { return false }
func newPublicationPhase(ctx context.Context, m *publicationStorageManager, address string) (*BuildPhase, *buildImage.Image, stage.Interface) {
	conveyor := &Conveyor{werfConfig: &config.WerfConfig{Meta: &config.Meta{Project: "publication-project"}}, StorageManager: m, stageImages: make(map[string]*stage.StageImage), serviceRWMutex: make(map[string]*sync.RWMutex)}
	var err error
	conveyor.StorageLockManager, err = lock_manager.NewHttp(ctx, address, "shared-client-id")
	gomega.Expect(err).To(gomega.Succeed())
	phase := NewBuildPhase(conveyor, BuildPhaseOptions{})
	phase.StagesIterator = NewStagesIterator(conveyor)
	img := &buildImage.Image{Name: "app", TargetPlatform: "linux/amd64", CommonImageOptions: buildImage.CommonImageOptions{Conveyor: conveyor, ForceTargetPlatformLogging: true}}
	stg := &publicationStage{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{ImageName: "app"})}
	stg.SetDigest("shared-digest")
	stg.SetContentDigest("loser-content")
	stg.SetStageImage(&stage.StageImage{Image: &publicationImage{name: "temporary"}})
	conveyor.SetStageImage(&stage.StageImage{Image: &publicationImage{name: "repo:shared-digest-100"}})
	parent := stage.NewBaseStage(stage.Install, &stage.BaseStageOptions{})
	parent.SetDigest("different-parent")
	parent.SetStageImage(&stage.StageImage{Image: &publicationImage{name: "parent", stageDesc: &imagePkg.StageDesc{StageID: imagePkg.NewStageID("different-parent", 10), Info: &imagePkg.Info{}}}})
	phase.StagesIterator.PrevNonEmptyStage = parent
	phase.StagesIterator.PrevBuiltStage = parent
	return phase, img, stg
}

func newPublicationLockServer() *httptest.Server {
	backend := distributed_locker.NewOptimisticLockingStorageBasedBackend(optimistic_locking_store.NewInMemoryStore())
	srv := httptest.NewServer(http.StripPrefix("/shared-client-id/locker", distributed_locker.NewHttpBackendHandler(backend)))
	ginkgo.DeferCleanup(srv.Close)
	return srv
}

type buildableStage struct{ *publicationStage }

var _ stage.Interface = (*buildableStage)(nil)

func (s *buildableStage) IsBuildable() bool { return true }

type stageBuilderStub struct {
	stage_builder.StageBuilderInterface
	builds int
}

var _ stage_builder.StageBuilderInterface = (*stageBuilderStub)(nil)

func (b *stageBuilderStub) Build(_ context.Context, _ container_backend.BuildOptions) error {
	b.builds++
	return nil
}

func eventCounts(collector *opstats.Collector) map[opstats.Event]int {
	counts := make(map[opstats.Event]int)
	for _, e := range collector.EventSummary() {
		counts[e.Event] = e.Count
	}
	return counts
}
