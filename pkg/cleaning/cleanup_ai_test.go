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

	deleteImageErrs  map[string]error
	deleteRecordErrs map[string]error
	deleteTagErrs    map[string]error

	deletedImages  []image.StageID
	deletedRecords []image.StageID
	deletedTags    []string
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

type fakeStorageManager struct {
	manager.StorageManagerInterface

	stages *fakePrimaryStagesStorage
}

func newFakeStorageManager() *fakeStorageManager {
	return &fakeStorageManager{
		stages: &fakePrimaryStagesStorage{
			deleteImageErrs:  map[string]error{},
			deleteRecordErrs: map[string]error{},
			deleteTagErrs:    map[string]error{},
		},
	}
}

func (f *fakeStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return f.stages
}

func (f *fakeStorageManager) GetImagesStorage() storage.StagesStorage {
	return nil
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
