package manager

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
)

type routingFakeStorage struct {
	storage.PrimaryStagesStorage

	name    string
	address string
	mu      sync.Mutex

	rmImageMetadata     []string
	rmManagedImage      []string
	deleteCustomTag     []string
	unregisterCustomTag []string
	getCustomTagMeta    []string

	deleteCustomTagErr error
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
	return f.deleteCustomTagErr
}

func (f *routingFakeStorage) UnregisterStageCustomTag(_ context.Context, tag storage.CustomTag) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unregisterCustomTag = append(f.unregisterCustomTag, tag.String())
	return nil
}

func (f *routingFakeStorage) Address() string {
	return f.address
}

func (f *routingFakeStorage) GetStageCustomTagMetadata(_ context.Context, id string) (*storage.CustomTagMetadata, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getCustomTagMeta = append(f.getCustomTagMeta, id)
	return &storage.CustomTagMetadata{Tag: id}, nil
}

func TestMetaStorageRouting_MetadataOpsGoToMeta_AliasStaysOnStages(t *testing.T) {
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

	require.NoError(t, sm.ForEachDeleteStageCustomTag(ctx, []storage.CustomTag{{Tag: "tag1"}},
		func(_ context.Context, _ storage.CustomTag, err error) error { return err }))

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

func TestMetaStorageRouting_FallbackToStagesWhenMetaNil(t *testing.T) {
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

func TestDeleteStageCustomTag_DeletesFromTheRepoTheTagLivesIn(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages", address: "registry/app"}
	meta := &routingFakeStorage{name: "meta", address: "registry/app/meta"}
	final := &routingFakeStorage{name: "final", address: "registry/app/install"}
	sibling := &routingFakeStorage{name: "sibling", address: "registry/app/tests"}

	var created []string
	sm := &StorageManager{
		ProjectName:            "proj",
		StagesStorage:          stages,
		MetaStorage:            meta,
		FinalStagesStorageList: []storage.StagesStorage{final},
		NewStagesStorageByAddress: func(_ context.Context, address string) (storage.StagesStorage, error) {
			created = append(created, address)
			if address != sibling.address {
				return nil, errors.New("unexpected address")
			}
			return sibling, nil
		},
	}
	ctx := context.Background()

	tags := []storage.CustomTag{
		{Tag: "main"},
		{Repo: stages.address, Tag: "explicit-primary"},
		{Repo: final.address, Tag: "main"},
		{Repo: sibling.address, Tag: "main"},
		{Repo: sibling.address, Tag: "pr1"},
	}
	for _, tag := range tags {
		require.NoError(t, sm.DeleteStageCustomTag(ctx, tag))
	}

	assert.Equal(t, []string{"main", "explicit-primary"}, stages.deleteCustomTag, "tags without repo or with the primary repo are deleted from stages")
	assert.Equal(t, []string{"main"}, final.deleteCustomTag, "a tag placed in a configured final repo is deleted there")
	assert.Equal(t, []string{"main", "pr1"}, sibling.deleteCustomTag, "a tag placed in another repo is deleted from a storage created for that repo")
	assert.Equal(t, []string{sibling.address}, created, "the storage for an unknown repo is created once and reused")
	assert.Empty(t, meta.deleteCustomTag, "nothing is deleted from meta")

	assert.Equal(t, []string{
		"main",
		stages.address + ":explicit-primary",
		final.address + ":main",
		sibling.address + ":main",
		sibling.address + ":pr1",
	}, meta.unregisterCustomTag, "every record is unregistered in meta with its repo")
}

func TestDeleteStageCustomTag_UnknownRepoWithoutFactoryFails(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages", address: "registry/app"}
	sm := &StorageManager{ProjectName: "proj", StagesStorage: stages}

	err := sm.DeleteStageCustomTag(context.Background(), storage.CustomTag{Repo: "registry/app/tests", Tag: "main"})
	require.ErrorIs(t, err, ErrCustomTagStorageUnavailable)
	assert.Empty(t, stages.deleteCustomTag, "a tag of another repo must never be deleted from the primary repo")
	assert.Empty(t, stages.unregisterCustomTag)
}

func TestDeleteStageCustomTag_FactoryFailureIsUnavailableStorage(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages", address: "registry/app"}
	sm := &StorageManager{
		ProjectName:   "proj",
		StagesStorage: stages,
		NewStagesStorageByAddress: func(_ context.Context, _ string) (storage.StagesStorage, error) {
			return nil, errors.New("no credentials")
		},
	}

	err := sm.DeleteStageCustomTag(context.Background(), storage.CustomTag{Repo: "registry/app/tests", Tag: "main"})
	require.ErrorIs(t, err, ErrCustomTagStorageUnavailable)
	assert.Empty(t, stages.deleteCustomTag)
}

func TestDeleteStageCustomTag_EquivalentRepoSpellingsResolveToPrimary(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages", address: "docker.io/acme/app"}
	meta := &routingFakeStorage{name: "meta", address: "docker.io/acme/meta"}
	sm := &StorageManager{ProjectName: "proj", StagesStorage: stages, MetaStorage: meta}

	require.NoError(t, sm.DeleteStageCustomTag(context.Background(), storage.CustomTag{Repo: "index.docker.io/acme/app", Tag: "main"}))
	assert.Equal(t, []string{"main"}, stages.deleteCustomTag, "another spelling of the primary repo is the primary repo")
}

func TestDeleteStageCustomTag_AliasDeleteFailureKeepsMetadata(t *testing.T) {
	require.NoError(t, werf.Init(t.TempDir(), ""))

	stages := &routingFakeStorage{name: "stages", address: "registry/app", deleteCustomTagErr: errors.New("registry down")}
	meta := &routingFakeStorage{name: "meta", address: "registry/app/meta"}
	sm := &StorageManager{ProjectName: "proj", StagesStorage: stages, MetaStorage: meta}

	err := sm.DeleteStageCustomTag(context.Background(), storage.CustomTag{Tag: "main"})
	require.Error(t, err)
	assert.Equal(t, []string{"main"}, stages.deleteCustomTag, "deletion was attempted")
	assert.Empty(t, meta.unregisterCustomTag, "the record must stay so the next cleanup retries the alias")
}
