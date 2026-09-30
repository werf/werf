package build

import (
	"context"
	"os"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/container_backend"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/manager"
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
