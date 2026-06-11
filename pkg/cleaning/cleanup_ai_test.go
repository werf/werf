package cleaning

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type fakePrimaryStagesStorage struct {
	storage.PrimaryStagesStorage

	rejectedStageIDs []image.StageID
	rejectedErr      error
}

func (f *fakePrimaryStagesStorage) GetRejectedStageIDs(_ context.Context, _ ...storage.Option) ([]image.StageID, error) {
	return f.rejectedStageIDs, f.rejectedErr
}

type fakeStorageManager struct {
	manager.StorageManagerInterface

	stages *fakePrimaryStagesStorage

	mu                  sync.Mutex
	deletedRejected     []image.StageID
	deleteRejectedErrs  map[string]error
	deletedCustomTags   []string
	deleteCustomTagErrs map[string]error
}

func newFakeStorageManager() *fakeStorageManager {
	return &fakeStorageManager{
		stages:              &fakePrimaryStagesStorage{},
		deleteRejectedErrs:  map[string]error{},
		deleteCustomTagErrs: map[string]error{},
	}
}

func (f *fakeStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return f.stages
}

func (f *fakeStorageManager) ForEachDeleteRejectedStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stageIDs []image.StageID, cb func(ctx context.Context, stageID image.StageID, err error) error) error {
	for _, id := range stageIDs {
		f.mu.Lock()
		f.deletedRejected = append(f.deletedRejected, id)
		err := f.deleteRejectedErrs[id.String()]
		f.mu.Unlock()
		if cbErr := cb(ctx, id, err); cbErr != nil {
			return cbErr
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachDeleteStageCustomTag(ctx context.Context, ids []string, cb func(ctx context.Context, tag string, err error) error) error {
	for _, id := range ids {
		f.mu.Lock()
		f.deletedCustomTags = append(f.deletedCustomTags, id)
		err := f.deleteCustomTagErrs[id]
		f.mu.Unlock()
		if cbErr := cb(ctx, id, err); cbErr != nil {
			return cbErr
		}
	}
	return nil
}

func TestAI_cleanupRejectedStages_NoRejected(t *testing.T) {
	sm := newFakeStorageManager()

	deleted, err := cleanupRejectedStages(context.Background(), sm, nil, nil, false)
	require.NoError(t, err)
	assert.Empty(t, deleted)
	assert.Empty(t, sm.deletedRejected)
	assert.Empty(t, sm.deletedCustomTags)
}

func TestAI_cleanupRejectedStages_DeletesLinkedCustomTags(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)
	otherStageID := image.NewStageID(digest, 1700000999)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	customTagsByStageID := map[string][]string{
		stageID.String():      {"v1.0.0", "latest"},
		otherStageID.String(): {"unrelated"},
	}

	deleted, err := cleanupRejectedStages(context.Background(), sm, customTagsByStageID, nil, false)
	require.NoError(t, err)

	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Equal(t, []image.StageID{*stageID}, sm.deletedRejected)
	assert.ElementsMatch(t, []string{"v1.0.0", "latest"}, sm.deletedCustomTags, "must delete linked custom tags, untouched 'unrelated'")
}

func TestAI_cleanupRejectedStages_DryRun(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	deleted, err := cleanupRejectedStages(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0"}}, nil, true)
	require.NoError(t, err)
	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Empty(t, sm.deletedRejected, "dry run must not touch registry")
	assert.Empty(t, sm.deletedCustomTags, "dry run must not delete custom tags")
}

func TestAI_cleanupRejectedStages_PropagatesGetError(t *testing.T) {
	sm := newFakeStorageManager()
	sm.stages.rejectedErr = errors.New("registry down")

	_, err := cleanupRejectedStages(context.Background(), sm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get rejected stage ids")
}

func TestAI_cleanupRejectedStages_DeletionErrorSwallowed(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.deleteRejectedErrs[stageID.String()] = errors.New("temporary network glitch")

	deleted, err := cleanupRejectedStages(context.Background(), sm, nil, nil, false)
	require.NoError(t, err, "non-fatal deletion errors must be logged but not fail cleanup")
	assert.Empty(t, deleted, "failed deletions must not appear in the returned 'successfully deleted' list")
}

func TestAI_cleanupRejectedStages_FatalDeletionErrorPropagates(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.deleteRejectedErrs[stageID.String()] = errors.New("UNAUTHORIZED")

	_, err := cleanupRejectedStages(context.Background(), sm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNAUTHORIZED")
}

func TestAI_cleanupRejectedStages_SkipsProtectedStages(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	protectedID := image.NewStageID(digest, 1700000000)
	freeID := image.NewStageID(digest, 1700000001)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*protectedID, *freeID}

	customTagsByStageID := map[string][]string{
		protectedID.String(): {"deployed-tag"},
		freeID.String():      {"orphan-tag"},
	}
	protected := map[string]bool{protectedID.String(): true}

	deleted, err := cleanupRejectedStages(context.Background(), sm, customTagsByStageID, protected, false)
	require.NoError(t, err)

	assert.Equal(t, []string{freeID.String()}, deleted, "protected stage must not be reported as deleted")
	assert.Equal(t, []image.StageID{*freeID}, sm.deletedRejected, "protected stage must not be deleted from registry")
	assert.Equal(t, []string{"orphan-tag"}, sm.deletedCustomTags, "custom tag of protected stage must NOT be deleted")
}

func TestAI_cleanupRejectedStages_AllProtectedNoop(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	protected := map[string]bool{stageID.String(): true}

	deleted, err := cleanupRejectedStages(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0"}}, protected, false)
	require.NoError(t, err)
	assert.Empty(t, deleted)
	assert.Empty(t, sm.deletedRejected, "registry must not be touched when every rejected stage is protected")
	assert.Empty(t, sm.deletedCustomTags, "linked custom tags must NOT be deleted when their parent is protected")
}
