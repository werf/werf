package manager

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

type routingFakeStorage struct {
	storage.PrimaryStagesStorage

	name string
	mu   sync.Mutex

	rmImageMetadata     []string
	rmManagedImage      []string
	deleteCustomTag     []string
	unregisterCustomTag []string
	getCustomTagMeta    []string
}

func (f *routingFakeStorage) RmImageMetadata(_ context.Context, _, imageNameOrID, commit, stageID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rmImageMetadata = append(f.rmImageMetadata, imageNameOrID+"/"+commit+"/"+stageID)
	return nil
}

func (f *routingFakeStorage) RmManagedImage(_ context.Context, _, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rmManagedImage = append(f.rmManagedImage, name)
	return nil
}

func (f *routingFakeStorage) DeleteStageCustomTag(_ context.Context, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteCustomTag = append(f.deleteCustomTag, tag)
	return nil
}

func (f *routingFakeStorage) UnregisterStageCustomTag(_ context.Context, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unregisterCustomTag = append(f.unregisterCustomTag, tag)
	return nil
}

func (f *routingFakeStorage) GetStageCustomTagMetadata(_ context.Context, id string) (*storage.CustomTagMetadata, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getCustomTagMeta = append(f.getCustomTagMeta, id)
	return &storage.CustomTagMetadata{Tag: id}, nil
}

func TestAI_MetaStorageRouting_MetadataOpsGoToMeta_AliasStaysOnStages(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages"}
	meta := &routingFakeStorage{name: "meta"}

	sm := &StorageManager{
		ProjectName:   "proj",
		StagesStorage: stages,
		MetaStorage:   meta,
	}
	ctx := context.Background()

	require.NoError(t, sm.ForEachRmImageMetadata(ctx, "proj", "img1", map[string][]string{
		"stage-id-1": {"commit-a"},
	}, func(_ context.Context, _, _ string, err error) error { return err }))

	require.NoError(t, sm.ForEachRmManagedImage(ctx, "proj", []string{"mng1", "mng2"},
		func(_ context.Context, _ string, err error) error { return err }))

	require.NoError(t, sm.ForEachDeleteStageCustomTag(ctx, []string{"tag1"},
		func(_ context.Context, _ string, err error) error { return err }))

	require.NoError(t, sm.ForEachGetStageCustomTagMetadata(ctx, []string{"id1"},
		func(_ context.Context, _ string, _ *storage.CustomTagMetadata, err error) error { return err }))

	assert.Equal(t, []string{"img1/commit-a/stage-id-1"}, meta.rmImageMetadata, "RmImageMetadata MUST go to meta")
	assert.Empty(t, stages.rmImageMetadata, "RmImageMetadata MUST NOT go to stages")

	assert.ElementsMatch(t, []string{"mng1", "mng2"}, meta.rmManagedImage, "RmManagedImage MUST go to meta")
	assert.Empty(t, stages.rmManagedImage, "RmManagedImage MUST NOT go to stages")

	assert.Equal(t, []string{"tag1"}, stages.deleteCustomTag, "DeleteStageCustomTag (alias image) MUST stay on stages")
	assert.Empty(t, meta.deleteCustomTag, "DeleteStageCustomTag MUST NOT go to meta")

	assert.Equal(t, []string{"tag1"}, meta.unregisterCustomTag, "UnregisterStageCustomTag (metadata record) MUST go to meta")
	assert.Empty(t, stages.unregisterCustomTag, "UnregisterStageCustomTag MUST NOT go to stages")

	assert.Equal(t, []string{"id1"}, meta.getCustomTagMeta, "GetStageCustomTagMetadata MUST go to meta")
	assert.Empty(t, stages.getCustomTagMeta, "GetStageCustomTagMetadata MUST NOT go to stages")
}

func TestAI_MetaStorageRouting_FallbackToStagesWhenMetaNil(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages"}
	sm := &StorageManager{
		ProjectName:   "proj",
		StagesStorage: stages,
	}

	assert.Same(t, stages, sm.GetMetaStorage(), "GetMetaStorage MUST fall back to StagesStorage when MetaStorage is nil")

	ctx := context.Background()
	require.NoError(t, sm.ForEachRmManagedImage(ctx, "proj", []string{"mng1"},
		func(_ context.Context, _ string, err error) error { return err }))

	assert.Equal(t, []string{"mng1"}, stages.rmManagedImage, "without meta, RmManagedImage falls through to stages")
}
