package build

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/lockgate/pkg/distributed_locker/optimistic_locking_store"
	buildImage "github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/container_backend"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/manager"
	"github.com/werf/werf/v3/pkg/storage/synchronization/lock_manager"
)

func newImage(name string, baseImageType buildImage.BaseImageType, opts buildImage.ImageOptions) *buildImage.Image {
	img, err := buildImage.NewImage(context.Background(), "linux/amd64", name, baseImageType, opts)
	gomega.Expect(err).To(gomega.Succeed())
	return img
}

func writeBuildReport(records ...ReportImageRecord) string {
	report := NewImagesReport()
	for _, record := range records {
		report.SetImageRecord(record.WerfImageName, record)
	}

	data, err := report.ToJsonData()
	gomega.Expect(err).To(gomega.Succeed())

	path := filepath.Join(ginkgo.GinkgoT().TempDir(), "report.json")
	gomega.Expect(os.WriteFile(path, data, 0o644)).To(gomega.Succeed())

	return path
}

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

type contentDependenciesStub struct {
	*stage.BaseStage
	deps string
	err  error
}

var _ stage.Interface = (*contentDependenciesStub)(nil)

func newContentDependenciesStub(name stage.StageName, deps string) *contentDependenciesStub {
	return &contentDependenciesStub{BaseStage: stage.NewBaseStage(name, &stage.BaseStageOptions{}), deps: deps}
}

func (s *contentDependenciesStub) GetContentDependencies(_ context.Context, _ stage.Conveyor, _ container_backend.BuildContextArchiver) (string, error) {
	return s.deps, s.err
}

var _ storage.PrimaryStagesStorage = (*anchorPrimaryStagesStorage)(nil)

type anchorPrimaryStagesStorage struct {
	storage.PrimaryStagesStorage
}

func (m *anchorLookupStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return m.primaryStagesStorage
}

var _ storage.PrimaryStagesStorage = (*exportStagesStorageStub)(nil)

type exportStagesStorageStub struct {
	storage.PrimaryStagesStorage
	address  string
	exported map[string]*imagePkg.StageDesc
}

func newExportStagesStorageStub(address string) *exportStagesStorageStub {
	return &exportStagesStorageStub{address: address, exported: map[string]*imagePkg.StageDesc{}}
}

func (s *exportStagesStorageStub) Address() string {
	return s.address
}

func (s *exportStagesStorageStub) ExportStage(_ context.Context, stageDesc *imagePkg.StageDesc, destinationReference string, _ func(config v1.Config) (v1.Config, error)) error {
	s.exported[destinationReference] = stageDesc
	return nil
}

var _ manager.StorageManagerInterface = (*exportStorageManager)(nil)

type exportStorageManager struct {
	manager.StorageManagerInterface
	stagesStorage      storage.PrimaryStagesStorage
	finalStagesStorage storage.StagesStorage
	finalStageDesc     *imagePkg.StageDesc
	copyOptions        manager.CopyStageIntoStorageOptions
}

func (m *exportStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return m.stagesStorage
}

func (m *exportStorageManager) GetFinalStagesStorage() storage.StagesStorage {
	return m.finalStagesStorage
}

func (m *exportStorageManager) CopyStageIntoFinalStorage(_ context.Context, _ imagePkg.StageID, _ storage.StagesStorage, opts manager.CopyStageIntoStorageOptions) (*imagePkg.StageDesc, error) {
	m.copyOptions = opts
	return m.finalStageDesc, nil
}

var _ container_backend.LegacyImageInterface = (*exportLegacyImageStub)(nil)

type exportLegacyImageStub struct {
	container_backend.LegacyImageInterface
	stageDesc *imagePkg.StageDesc
}

func (i *exportLegacyImageStub) GetStageDesc() *imagePkg.StageDesc {
	return i.stageDesc
}

type introspectPipelineRecord struct {
	processedStages []string
	introspected    []string
}

var _ container_backend.LegacyImageInterface = (*introspectImageStub)(nil)

type introspectImageStub struct {
	container_backend.LegacyImageInterface
	name      string
	stageDesc *imagePkg.StageDesc
	record    *introspectPipelineRecord
}

func (i *introspectImageStub) Name() string              { return i.name }
func (i *introspectImageStub) GetTargetPlatform() string { return "linux/amd64" }

func (i *introspectImageStub) SetStageDesc(desc *imagePkg.StageDesc) { i.stageDesc = desc }
func (i *introspectImageStub) GetStageDesc() *imagePkg.StageDesc     { return i.stageDesc }

func (i *introspectImageStub) Introspect(_ context.Context) error {
	i.record.introspected = append(i.record.introspected, i.name)
	return nil
}

var _ stage.Interface = (*introspectStageStub)(nil)

type introspectStageStub struct {
	*stage.BaseStage
	record *introspectPipelineRecord
}

func newIntrospectStageStub(name stage.StageName, imageName string, isContentAnchor bool, record *introspectPipelineRecord) *introspectStageStub {
	base := stage.NewBaseStage(name, &stage.BaseStageOptions{ImageName: imageName})
	base.SetContentAnchor(isContentAnchor)
	return &introspectStageStub{BaseStage: base, record: record}
}

func (s *introspectStageStub) HasPrevStage() bool { return false }

func (s *introspectStageStub) GetDependencies(_ context.Context, _ stage.Conveyor, _ container_backend.ContainerBackend, _, _ *stage.StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	return "", nil
}

func (s *introspectStageStub) IsEmpty(_ context.Context, _ stage.Conveyor, _ *stage.StageImage) (bool, error) {
	s.record.processedStages = append(s.record.processedStages, string(s.Name()))
	return false, nil
}

func newIntrospectStageDesc(name string) *imagePkg.StageDesc {
	return &imagePkg.StageDesc{
		StageID: imagePkg.NewStageID("digest", 1),
		Info: &imagePkg.Info{
			Name:   name,
			Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "content"},
		},
	}
}

func (m *anchorLookupStorageManager) GetStageDescSetByDigestFromStagesStorageCached(_ context.Context, _, _ string, _ int64, stagesStorage storage.StagesStorage) (imagePkg.StageDescSet, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if stagesStorage == m.secondaryStagesStorage {
		m.cachedSecondaryLookups++
		return m.inSecondary, nil
	}

	m.cachedPrimaryLookups++
	return m.inPrimary, nil
}

type publicationStorage struct {
	storage.PrimaryStagesStorage
	mutex          sync.Mutex
	desc           *imagePkg.StageDesc
	writes         int
	writeErr       error
	descriptionErr error
}

var _ storage.PrimaryStagesStorage = (*publicationStorage)(nil)

func (s *publicationStorage) String() string { return "publication-test" }
func (s *publicationStorage) StoreImage(_ context.Context, img container_backend.LegacyImageInterface) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.writeErr != nil {
		return s.writeErr
	}
	s.writes++
	s.desc = &imagePkg.StageDesc{StageID: imagePkg.NewStageID("shared-digest", 100), Info: &imagePkg.Info{Name: img.Name(), Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "winner-content"}}}
	return nil
}

func (s *publicationStorage) GetStageDesc(_ context.Context, _ string, _ imagePkg.StageID) (*imagePkg.StageDesc, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.desc, s.descriptionErr
}

type publicationStorageManager struct {
	manager.StorageManagerInterface
	primary        *publicationStorage
	lookups        atomic.Int32
	lookupStarted  chan struct{}
	continueLookup chan struct{}
	lookupErr      error
	selectionErr   error
	cacheErr       error
	parentTs       int64
	secondary      *publicationStorage
	secondaryDesc  *imagePkg.StageDesc
	copies         atomic.Int32
}

var _ manager.StorageManagerInterface = (*publicationStorageManager)(nil)

func (m *publicationStorageManager) GetStagesStorage() storage.PrimaryStagesStorage { return m.primary }
func (m *publicationStorageManager) GetStageDescSetByDigest(_ context.Context, _, _ string, parentTs int64) (imagePkg.StageDescSet, error) {
	m.parentTs = parentTs
	m.lookups.Add(1)
	m.primary.mutex.Lock()
	set := imagePkg.NewStageDescSet()
	if m.primary.desc != nil && m.primary.desc.StageID.CreationTs >= parentTs {
		set = imagePkg.NewStageDescSet(m.primary.desc)
	}
	m.primary.mutex.Unlock()
	if m.lookupStarted != nil {
		close(m.lookupStarted)
		<-m.continueLookup
	}
	return set, m.lookupErr
}

func (m *publicationStorageManager) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, stg stage.Interface, set imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	if m.selectionErr != nil {
		return nil, m.selectionErr
	}
	return stg.SelectSuitableStageDesc(ctx, c, set)
}

func (m *publicationStorageManager) GenerateStageDescCreationTs(_ string, _ imagePkg.StageDescSet) (string, int64) {
	return "repo:shared-digest-100", 100
}
func (m *publicationStorageManager) GetCacheStagesStorageList() []storage.StagesStorage { return nil }
func (m *publicationStorageManager) CopyStageIntoCacheStorages(_ context.Context, _ imagePkg.StageID, _ []storage.StagesStorage, _ manager.CopyStageIntoStorageOptions) error {
	return m.cacheErr
}

type publicationStage struct{ *stage.BaseStage }

var _ stage.Interface = (*publicationStage)(nil)

func (s *publicationStage) IsBuildable() bool { return false }

type publicationImage struct{ *introspectImageStub }

var _ container_backend.LegacyImageInterface = (*publicationImage)(nil)

func (i *publicationImage) SetName(name string) { i.name = name }

func newPublicationPhase(ctx context.Context, m *publicationStorageManager, address string, anchor bool, parentTs int64) (*BuildPhase, *buildImage.Image, stage.Interface) {
	phase := newTestBuildPhase(m, nil)
	phase.Conveyor.werfConfig = &config.WerfConfig{Meta: &config.Meta{Project: "publication-project"}}
	var err error
	phase.Conveyor.StorageLockManager, err = lock_manager.NewHttp(ctx, address, "shared-client-id")
	gomega.Expect(err).To(gomega.Succeed())
	phase.StagesIterator = NewStagesIterator(phase.Conveyor)
	img := newTestImage("app", true)
	img.Conveyor = phase.Conveyor
	img.ForceTargetPlatformLogging = true
	stg := &publicationStage{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{ImageName: "app"})}
	stg.SetContentAnchor(anchor)
	stg.SetDigest("shared-digest")
	stg.SetContentDigest("loser-content")
	stg.SetStageImage(&stage.StageImage{Image: &publicationImage{introspectImageStub: &introspectImageStub{name: "temporary", record: &introspectPipelineRecord{}}}})
	phase.Conveyor.SetStageImage(&stage.StageImage{Image: &publicationImage{introspectImageStub: &introspectImageStub{name: "repo:shared-digest-100"}}})
	phase.Conveyor.SetStageImage(&stage.StageImage{Image: &publicationImage{introspectImageStub: &introspectImageStub{name: "repo:alternate"}}})
	parent := stage.NewBaseStage(stage.Install, &stage.BaseStageOptions{})
	parent.SetDigest(fmt.Sprintf("different-parent-%d", parentTs))
	parent.SetStageImage(&stage.StageImage{Image: &publicationImage{introspectImageStub: &introspectImageStub{name: "parent", stageDesc: &imagePkg.StageDesc{StageID: imagePkg.NewStageID(parent.GetDigest(), parentTs), Info: &imagePkg.Info{}}}}})
	phase.StagesIterator.PrevNonEmptyStage = parent
	phase.StagesIterator.PrevBuiltStage = parent
	return phase, img, stg
}

func newPublicationLockServer() (*httptest.Server, <-chan struct{}) {
	backend := distributed_locker.NewOptimisticLockingStorageBasedBackend(optimistic_locking_store.NewInMemoryStore())
	handler := http.StripPrefix("/shared-client-id/locker", distributed_locker.NewHttpBackendHandler(backend))
	attempts := make(chan struct{}, 16)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		if r.URL.Path == "/shared-client-id/locker/acquire" {
			attempts <- struct{}{}
		}
	}))
	ginkgo.DeferCleanup(srv.Close)
	return srv, attempts
}

func (m *publicationStorageManager) GetSecondaryStagesStorageList() []storage.StagesStorage {
	return []storage.StagesStorage{m.secondary}
}

func (m *publicationStorageManager) GetStageDescSetByDigestFromStagesStorageWithCache(_ context.Context, _, _ string, _ int64, _ storage.StagesStorage) (imagePkg.StageDescSet, error) {
	return imagePkg.NewStageDescSet(m.secondaryDesc), nil
}

func (m *publicationStorageManager) CopySuitableStageDescByDigest(_ context.Context, desc *imagePkg.StageDesc, _, _ storage.StagesStorage, _ container_backend.ContainerBackend, _ string) (*imagePkg.StageDesc, error) {
	m.copies.Add(1)
	m.primary.mutex.Lock()
	defer m.primary.mutex.Unlock()
	m.primary.writes++
	m.primary.desc = desc
	return desc, nil
}

func (m *publicationStorageManager) GetStageDescSetByDigestWithCache(_ context.Context, _, _ string, _ int64) (imagePkg.StageDescSet, error) {
	m.lookups.Add(1)
	if m.lookupStarted != nil {
		close(m.lookupStarted)
		<-m.continueLookup
	}
	return imagePkg.NewStageDescSet(), nil
}
