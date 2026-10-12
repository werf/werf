package cleaning

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/cleaning/stage_manager"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

var _ manager.StorageManagerInterface = (*fakeStorageManager)(nil)

func newCleanupManagerForImportsMetadataTest(ctx context.Context, protectedStageDescs ...*image.StageDesc) *cleanupManager {
	ginkgo.GinkgoHelper()
	sm := newFakeStorageManager()
	sm.stageDescSet = image.NewStageDescSet(protectedStageDescs...)

	stageManager := stage_manager.NewManager()
	gomega.Expect(stageManager.InitStageDescSet(ctx, sm)).To(gomega.Succeed())
	for _, stageDesc := range protectedStageDescs {
		stageManager.MarkStageDescAsProtected(stageDesc, stage_manager.ProtectionReasonImportSource, false)
	}

	return &cleanupManager{
		stageManager:   stageManager,
		StorageManager: sm,
		ProjectName:    "myproject",
		report:         newTestReport(),
	}
}

func newTestStageDesc(repository string, stageID *image.StageID) *image.StageDesc {
	return &image.StageDesc{
		StageID: stageID,
		Info: &image.Info{
			Repository: repository,
			Tag:        stageID.String(),
			Name:       repository + ":" + stageID.String(),
		},
	}
}

func (f *fakeStorageManager) ForEachDeleteFinalStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stages image.StageDescSet, onDelete func(context.Context, *image.StageDesc, error) error) error {
	for stage := range stages.Iter() {
		f.deletedFinalStages = append(f.deletedFinalStages, stage)
		if err := onDelete(ctx, stage, nil); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachDeleteStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stages image.StageDescSet, onDelete func(context.Context, *image.StageDesc, error) error) error {
	for stage := range stages.Iter() {
		f.deletedStages = append(f.deletedStages, stage)
		if err := onDelete(ctx, stage, nil); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachGetStageCustomTagMetadata(_ context.Context, _ []string, _ func(context.Context, string, *storage.CustomTagMetadata, error) error) error {
	return nil
}

func (f *fakeStorageManager) GetFinalStagesStorage() storage.StagesStorage {
	return nil
}

func (f *fakePrimaryStagesStorage) GetAllAndGroupImageMetadataByImageName(_ context.Context, _ string, _ []string, _ ...storage.Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return nil, nil, nil
}

func (f *fakePrimaryStagesStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...storage.Option) ([]string, error) {
	return nil, nil
}

func (f *fakePrimaryStagesStorage) PostLastCleanupRecord(_ context.Context, projectName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastCleanupRecords = append(f.lastCleanupRecords, projectName)
	return nil
}

var _ storage.PrimaryStagesStorage = (*fakePrimaryStagesStorage)(nil)

func (f *fakePrimaryStagesStorage) UnregisterStageCustomTag(_ context.Context, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unregisteredTags = append(f.unregisteredTags, tag)
	return nil
}

func (f *fakeStorageManager) GetMetaStorage(_ context.Context) storage.PrimaryStagesStorage {
	if f.meta != nil {
		return f.meta
	}
	return f.stages
}

func newFakeStorageManagerWithSplitStorages() *fakeStorageManager {
	f := newFakeStorageManager()
	f.stages.address = "example.com/stages"
	f.meta = &fakePrimaryStagesStorage{address: "example.com/meta"}
	return f
}

var _ docker_registry.Interface = (*purgeMarkerRegistry)(nil)

type purgeMarkerRegistry struct {
	docker_registry.Interface
	mu                 sync.Mutex
	images             map[string]*image.Info
	failMetadataDelete bool
	deletionFailures   int
}

func (r *purgeMarkerRegistry) PushImage(_ context.Context, name string, opts *docker_registry.PushImageOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.images[name] = &image.Info{Name: name, Labels: opts.Labels}
	return nil
}

func (r *purgeMarkerRegistry) TryGetRepoImage(_ context.Context, name string) (*image.Info, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.images[name], nil
}

func (r *purgeMarkerRegistry) GetRepoImage(ctx context.Context, name string) (*image.Info, error) {
	return r.TryGetRepoImage(ctx, name)
}

func (r *purgeMarkerRegistry) Tags(_ context.Context, repo string, _ ...docker_registry.Option) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var tags []string
	for name := range r.images {
		if strings.HasPrefix(name, repo+":") {
			tags = append(tags, strings.TrimPrefix(name, repo+":"))
		}
	}
	return tags, nil
}

func (r *purgeMarkerRegistry) DeleteRepoImage(_ context.Context, info *image.Info) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failMetadataDelete && strings.HasPrefix(info.Name, "example.com/meta:") {
		r.deletionFailures++
		return fmt.Errorf("registry deletion temporarily unavailable")
	}
	delete(r.images, info.Name)
	return nil
}

func (f *fakePrimaryStagesStorage) Address() string { return f.address }
