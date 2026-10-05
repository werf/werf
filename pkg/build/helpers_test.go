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
	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/stage_builder"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
)

func newCachedLookupPhase(ctx context.Context) (*BuildPhase, *image.Image, *cachedLookupStage, *cachedLookupStorageManager) {
	storageManager := &cachedLookupStorageManager{}
	c := &Conveyor{
		StorageManager:     storageManager,
		StorageLockManager: storageManager,
		werfConfig:         &config.WerfConfig{Meta: &config.Meta{Project: "lookup-test"}},
		serviceRWMutex:     map[string]*sync.RWMutex{},
		stageDigestMutex:   map[string]*sync.Mutex{},
		stageImages:        map[string]*stage.StageImage{},
	}
	phase := NewBuildPhase(c, BuildPhaseOptions{})
	phase.StagesIterator = NewStagesIterator(c)
	img, err := image.NewImage(ctx, "linux/amd64", "app", image.NoBaseImage, image.ImageOptions{CommonImageOptions: image.CommonImageOptions{Conveyor: c, ForceTargetPlatformLogging: true}})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	stg := &cachedLookupStage{BaseStage: stage.NewBaseStage(stage.From, &stage.BaseStageOptions{ImageName: "app"})}
	return phase, img, stg, storageManager
}

func cachedLookupDesc(name, content string) *imagePkg.StageDesc {
	return &imagePkg.StageDesc{StageID: imagePkg.NewStageID("digest", 1), Info: &imagePkg.Info{Name: "repo:" + name, Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: content}}}
}

var _ stage.Interface = (*cachedLookupStage)(nil)

type cachedLookupStage struct{ *stage.BaseStage }

func (s *cachedLookupStage) HasPrevStage() bool { return false }

func (s *cachedLookupStage) GetDependencies(context.Context, stage.Conveyor, container_backend.ContainerBackend, *stage.StageImage, *stage.StageImage, container_backend.BuildContextArchiver) (string, error) {
	return "dependencies", nil
}

var (
	_ manager.StorageManagerInterface = (*cachedLookupStorageManager)(nil)
	_ lock_manager.Interface          = (*cachedLookupStorageManager)(nil)
	_ storage.PrimaryStagesStorage    = (*cachedLookupStorageManager)(nil)
)

type cachedLookupStorageManager struct {
	manager.StorageManagerInterface
	storage.PrimaryStagesStorage
	primary, secondary, publishOnLock                  *imagePkg.StageDesc
	cachedLookups, strictLookups, freshLookups, copies int
	locked                                             bool
}

func (m *cachedLookupStorageManager) GetStagesStorage() storage.PrimaryStagesStorage { return m }
func (m *cachedLookupStorageManager) GetSecondaryStagesStorageList() []storage.StagesStorage {
	return []storage.StagesStorage{m}
}
func (m *cachedLookupStorageManager) String() string { return "secondary" }

func (m *cachedLookupStorageManager) GetStageDescSetByDigestFromStagesStorageCached(context.Context, string, string, int64, storage.StagesStorage) (imagePkg.StageDescSet, error) {
	m.cachedLookups++
	if m.secondary != nil {
		return imagePkg.NewStageDescSet(m.secondary), nil
	}
	return imagePkg.NewStageDescSet(), nil
}

func (m *cachedLookupStorageManager) GetStageDescSetByDigestWithCache(context.Context, string, string, int64) (imagePkg.StageDescSet, error) {
	m.strictLookups++
	return imagePkg.NewStageDescSet(), nil
}

func (m *cachedLookupStorageManager) GetStageDescSetByDigest(context.Context, string, string, int64) (imagePkg.StageDescSet, error) {
	m.freshLookups++
	gomega.Expect(m.locked).To(gomega.BeTrue())
	if m.primary != nil {
		return imagePkg.NewStageDescSet(m.primary), nil
	}
	return imagePkg.NewStageDescSet(), nil
}

func (m *cachedLookupStorageManager) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, stg stage.Interface, set imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	return stg.SelectSuitableStageDesc(ctx, c, set)
}

func (m *cachedLookupStorageManager) CopySuitableStageDescByDigest(_ context.Context, desc *imagePkg.StageDesc, _, _ storage.StagesStorage, _ container_backend.ContainerBackend, _ string) (*imagePkg.StageDesc, error) {
	m.copies++
	return desc, nil
}

func (m *cachedLookupStorageManager) GetCacheStagesStorageList() []storage.StagesStorage { return nil }
func (m *cachedLookupStorageManager) CopyStageIntoCacheStorages(context.Context, imagePkg.StageID, []storage.StagesStorage, manager.CopyStageIntoStorageOptions) error {
	return nil
}

func (m *cachedLookupStorageManager) LockStage(context.Context, string, string) (lock_manager.LockHandle, error) {
	m.locked = true
	m.primary = m.publishOnLock
	return lock_manager.LockHandle{}, nil
}

func (m *cachedLookupStorageManager) Unlock(context.Context, lock_manager.LockHandle) error {
	m.locked = false
	return nil
}

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

func (m *checkModeStorageManager) GetFinalStagesStorage() storage.StagesStorage { return nil }

var _ storage.PrimaryStagesStorage = (*checkModeStorage)(nil)

type checkModeStorage struct {
	storage.PrimaryStagesStorage
	desc         *imagePkg.StageDesc
	writes       []string
	customTagErr error
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
	return s.customTagErr
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

var _ storage.PrimaryStagesStorage = (*importMetadataStorageStub)(nil)

type importMetadataStorageStub struct {
	storage.PrimaryStagesStorage
	metadata *storage.ImportMetadata
	err      error
	writes   int
}

func (s *importMetadataStorageStub) GetImportMetadata(context.Context, string, string) (*storage.ImportMetadata, error) {
	return s.metadata, s.err
}

func (s *importMetadataStorageStub) PutImportMetadata(context.Context, string, *storage.ImportMetadata, storage.PutImportMetadataOptions) error {
	s.writes++
	return nil
}

func newReportPhase(reportPath string) *BuildPhase {
	return NewBuildPhase(&Conveyor{}, BuildPhaseOptions{BuildOptions: BuildOptions{ReportPath: reportPath, ReportFormat: ReportJSON}})
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
func newPublicationPhase(ctx context.Context, m *publicationStorageManager, address string) (*BuildPhase, *image.Image, stage.Interface) {
	conveyor := &Conveyor{werfConfig: &config.WerfConfig{Meta: &config.Meta{Project: "publication-project"}}, StorageManager: m, stageImages: make(map[string]*stage.StageImage), serviceRWMutex: make(map[string]*sync.RWMutex)}
	var err error
	conveyor.StorageLockManager, err = lock_manager.NewHttp(ctx, address, "shared-client-id")
	gomega.Expect(err).To(gomega.Succeed())
	phase := NewBuildPhase(conveyor, BuildPhaseOptions{})
	phase.StagesIterator = NewStagesIterator(conveyor)
	img := &image.Image{Name: "app", TargetPlatform: "linux/amd64", CommonImageOptions: image.CommonImageOptions{Conveyor: conveyor, ForceTargetPlatformLogging: true}}
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
