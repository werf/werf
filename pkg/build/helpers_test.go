package build

import (
	"context"
	"sync"

	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	imagePkg "github.com/werf/werf/v2/pkg/image"
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

func (m *cachedLookupStorageManager) GetStageDescSetByDigestFromStagesStorageWithCache(context.Context, string, string, int64, storage.StagesStorage) (imagePkg.StageDescSet, error) {
	m.strictLookups++
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
