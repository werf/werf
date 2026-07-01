package stage_manager

import (
	"context"
	"sync"
	"testing"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type fakeMetaStorage struct {
	storage.PrimaryStagesStorage
	mu  sync.Mutex
	put []string // "image|commit|stageID"
}

func (f *fakeMetaStorage) GetAllAndGroupImageMetadataByImageName(_ context.Context, _ string, _ []string, _ ...storage.Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return map[string]map[string][]string{}, map[string]map[string][]string{}, nil
}

func (f *fakeMetaStorage) PutImageMetadata(_ context.Context, _, imageName, commit, stageID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.put = append(f.put, imageName+"|"+commit+"|"+stageID)
	return nil
}

type fakeSM struct {
	manager.StorageManagerInterface
	meta *fakeMetaStorage
}

func (f *fakeSM) GetMetaStagesStorage() storage.PrimaryStagesStorage { return f.meta }

type fakeGit struct{}

func (fakeGit) IsCommitExists(_ context.Context, _ string) (bool, error) { return true, nil }

// A degraded meta store (empty metadata, existing stages) must not delete
// anything: all stages get protected and the label commit is backfilled.
func TestInitImagesMetadata_DegradedProtectsAndBackfills(t *testing.T) {
	stageDesc := &image.StageDesc{
		StageID: image.NewStageID("digest-x", 1),
		Info: &image.Info{
			Labels: map[string]string{image.WerfProjectRepoCommitLabel: "commit-a"},
		},
	}
	m := NewManager()
	m.managedStageDescSet = newManagedStageDescSet(image.NewStageDescSet(stageDesc))

	sm := &fakeSM{meta: &fakeMetaStorage{}}

	if err := m.InitImagesMetadata(context.Background(), sm, fakeGit{}, "proj", []string{"app"}); err != nil {
		t.Fatal(err)
	}

	protected := m.managedStageDescSet.GetProtectedStageDescSet()
	if protected.Cardinality() != 1 {
		t.Fatalf("expected the stage to be protected in degraded mode, got %d protected", protected.Cardinality())
	}

	if len(sm.meta.put) != 1 || sm.meta.put[0] != "app|commit-a|"+stageDesc.StageID.String() {
		t.Fatalf("expected one backfilled record app|commit-a|<stageID>, got %v", sm.meta.put)
	}
}
