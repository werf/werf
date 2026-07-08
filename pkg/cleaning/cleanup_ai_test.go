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

	mu sync.Mutex

	rejectedStageIDs []image.StageID
	rejectedErr      error

	deleteImageErrs   map[string]error
	deleteRecordErrs  map[string]error
	deleteTagErrs     map[string]error
	unregisterTagErrs map[string]error

	deletedImages    []image.StageID
	deletedRecords   []image.StageID
	deletedTags      []string
	unregisteredTags []string
}

func (f *fakePrimaryStagesStorage) GetRejectedStageIDs(_ context.Context, _ ...storage.Option) ([]image.StageID, error) {
	return f.rejectedStageIDs, f.rejectedErr
}

func (f *fakePrimaryStagesStorage) DeleteRejectedStageImage(_ context.Context, stageID image.StageID, _ storage.DeleteImageOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedImages = append(f.deletedImages, stageID)
	return f.deleteImageErrs[stageID.String()]
}

func (f *fakePrimaryStagesStorage) DeleteRejectedStageRecord(_ context.Context, stageID image.StageID, _ storage.DeleteImageOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedRecords = append(f.deletedRecords, stageID)
	return f.deleteRecordErrs[stageID.String()]
}

func (f *fakePrimaryStagesStorage) DeleteStageCustomTag(_ context.Context, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedTags = append(f.deletedTags, tag)
	return f.deleteTagErrs[tag]
}

func (f *fakePrimaryStagesStorage) UnregisterStageCustomTag(_ context.Context, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unregisteredTags = append(f.unregisteredTags, tag)
	return f.unregisterTagErrs[tag]
}

type fakeStorageManager struct {
	manager.StorageManagerInterface

	stages *fakePrimaryStagesStorage
	meta   *fakePrimaryStagesStorage
}

func newFakeStorageManager() *fakeStorageManager {
	sm := &fakeStorageManager{
		stages: newFakePrimaryStagesStorage(),
	}
	sm.meta = sm.stages
	return sm
}

func newFakeStorageManagerWithSplitStorages() *fakeStorageManager {
	return &fakeStorageManager{
		stages: newFakePrimaryStagesStorage(),
		meta:   newFakePrimaryStagesStorage(),
	}
}

func newFakePrimaryStagesStorage() *fakePrimaryStagesStorage {
	return &fakePrimaryStagesStorage{
		deleteImageErrs:   map[string]error{},
		deleteRecordErrs:  map[string]error{},
		deleteTagErrs:     map[string]error{},
		unregisterTagErrs: map[string]error{},
	}
}

func (f *fakeStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return f.stages
}

func (f *fakeStorageManager) GetMetaStorage() storage.PrimaryStagesStorage {
	return f.meta
}

func (f *fakeStorageManager) ForEachRejectedStage(ctx context.Context, stageIDs []image.StageID, cb func(ctx context.Context, stageID image.StageID) error) error {
	for _, id := range stageIDs {
		if err := cb(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func TestAI_deleteRejectedStagesWithLinkedTags_NoRejected(t *testing.T) {
	sm := newFakeStorageManager()

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false)
	require.NoError(t, err)
	assert.Empty(t, deleted)
	assert.Empty(t, sm.stages.deletedImages)
	assert.Empty(t, sm.stages.deletedTags)
	assert.Empty(t, sm.stages.deletedRecords)
}

func TestAI_deleteRejectedStagesWithLinkedTags_OrderStageThenTagsThenMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)
	otherStageID := image.NewStageID(digest, 1700000999)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	customTagsByStageID := map[string][]string{
		stageID.String():      {"v1.0.0", "latest"},
		otherStageID.String(): {"unrelated"},
	}

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, customTagsByStageID, false)
	require.NoError(t, err)

	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages, "stage image deleted first")
	assert.Equal(t, []string{"v1.0.0", "latest"}, sm.stages.deletedTags, "linked custom tags deleted next, in given order")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedRecords, "marker deleted last")
}

func TestAI_deleteRejectedStagesWithLinkedTags_DryRun(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0"}}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Empty(t, sm.stages.deletedImages, "dry run must not touch registry")
	assert.Empty(t, sm.stages.deletedTags, "dry run must not touch registry")
	assert.Empty(t, sm.stages.deletedRecords, "dry run must not touch registry")
}

func TestAI_deleteRejectedStagesWithLinkedTags_PropagatesGetError(t *testing.T) {
	sm := newFakeStorageManager()
	sm.stages.rejectedErr = errors.New("registry down")

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get rejected stage ids")
}

func TestAI_deleteRejectedStagesWithLinkedTags_StageImageNonFatalFailureKeepsMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteImageErrs[stageID.String()] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0"}}, false)
	require.NoError(t, err)

	assert.Empty(t, deleted, "stage image deletion failed: stage not reported deleted, retry on next cleanup")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages, "attempt was made")
	assert.Empty(t, sm.stages.deletedTags, "custom tags must NOT be touched when stage image delete failed")
	assert.Empty(t, sm.stages.deletedRecords, "marker must remain so retry picks up this stage")
}

func TestAI_deleteRejectedStagesWithLinkedTags_StageImageFatalFailurePropagates(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteImageErrs[stageID.String()] = errors.New("UNAUTHORIZED")

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNAUTHORIZED")
}

func TestAI_deleteRejectedStagesWithLinkedTags_CustomTagFailureKeepsMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteTagErrs["v1.0.0"] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0", "latest"}}, false)
	require.NoError(t, err)

	assert.Empty(t, deleted, "stage with failed custom tag must NOT be reported deleted")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages, "stage image already deleted")
	assert.Equal(t, []string{"v1.0.0"}, sm.stages.deletedTags, "fail-fast on first custom tag failure; 'latest' not attempted")
	assert.Empty(t, sm.stages.deletedRecords, "marker MUST remain so next cleanup retries linked tags")
}

func TestAI_deleteRejectedStagesWithLinkedTags_RoutesUnregisterToMetaStorage(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManagerWithSplitStorages()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0", "latest"}}, false)
	require.NoError(t, err)

	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Equal(t, []string{"v1.0.0", "latest"}, sm.stages.deletedTags, "alias custom tags deleted from stages storage")
	assert.Equal(t, []string{"v1.0.0", "latest"}, sm.meta.unregisteredTags, "custom-tag metadata records unregistered from meta storage")
	assert.Empty(t, sm.meta.deletedTags, "meta storage MUST NOT receive alias image deletes")
	assert.Empty(t, sm.stages.unregisteredTags, "stages storage MUST NOT receive metadata unregister calls")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedRecords, "marker deleted on stages after both alias and metadata cleanup succeeded")
}

func TestAI_deleteRejectedStagesWithLinkedTags_UnregisterFailureKeepsMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManagerWithSplitStorages()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.meta.unregisterTagErrs["v1.0.0"] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0", "latest"}}, false)
	require.NoError(t, err)

	assert.Empty(t, deleted, "stage with failed metadata unregister must NOT be reported deleted")
	assert.Equal(t, []string{"v1.0.0"}, sm.stages.deletedTags, "alias for v1.0.0 already deleted before unregister failed")
	assert.Equal(t, []string{"v1.0.0"}, sm.meta.unregisteredTags, "fail-fast on first metadata unregister failure; latest not attempted")
	assert.Empty(t, sm.stages.deletedRecords, "marker MUST remain so next cleanup retries orphan metadata")
}

func TestAI_deleteRejectedStagesWithLinkedTags_MarkerFailureExcludesFromDeleted(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteRecordErrs[stageID.String()] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false)
	require.NoError(t, err)

	assert.Empty(t, deleted, "marker deletion failed: stage not in deleted list")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages)
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedRecords, "attempt was made")
}
