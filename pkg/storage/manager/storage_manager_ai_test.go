package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

func TestAI_GenerateStageDescCreationTs_AvoidsNameCollision(t *testing.T) {
	m := &StorageManager{
		ProjectName: "test-project",
		Storages:    Storages{Stages: storage.NewLocalRegistryStorage(nil)},
	}
	digest := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	_, baseTs := m.GenerateStageDescCreationTs(digest, image.NewStageDescSet())

	occupied := image.NewStageDescSet()
	occupiedNames := make(map[string]struct{})
	for ts := baseTs; ts <= baseTs+10; ts++ {
		name := m.Storages.Stages.ConstructStageImageName(m.ProjectName, digest, ts)
		occupied.Add(&image.StageDesc{
			StageID: image.NewStageID(digest, ts),
			Info:    &image.Info{Name: name},
		})
		occupiedNames[name] = struct{}{}
	}

	gotName, gotTs := m.GenerateStageDescCreationTs(digest, occupied)

	require.NotContains(t, occupiedNames, gotName, "must not collide with an occupied image name")
	require.Greater(t, gotTs, baseTs+10, "must increment past all occupied timestamps")
}

func TestAI_GetCacheStagesStorageList_ReturnsReadList(t *testing.T) {
	readStorage := storage.NewLocalRegistryStorage(nil)
	writeStorage := storage.NewLocalRegistryStorage(nil)
	m := &StorageManager{
		Storages: Storages{
			CacheFrom: []storage.RegistryStorage{readStorage},
			CacheTo:   []storage.RegistryStorage{writeStorage},
		},
	}

	got := m.GetCacheStagesStorageList()

	require.Len(t, got, 1)
	assert.Same(t, readStorage, got[0])
	assert.NotSame(t, writeStorage, got[0])
}

func TestAI_GetCacheStagesWriteList_ReturnsWriteList(t *testing.T) {
	readStorage := storage.NewLocalRegistryStorage(nil)
	writeStorage := storage.NewLocalRegistryStorage(nil)
	m := &StorageManager{
		Storages: Storages{
			CacheFrom: []storage.RegistryStorage{readStorage},
			CacheTo:   []storage.RegistryStorage{writeStorage},
		},
	}

	got := m.GetCacheStagesWriteList()

	require.Len(t, got, 1)
	assert.Same(t, writeStorage, got[0])
	assert.NotSame(t, readStorage, got[0])
}

func TestAI_IsRemoteImagesStorage_NilImagesIsNotRemote(t *testing.T) {
	s := Storages{}
	assert.False(t, s.IsRemoteImagesStorage())

	s.Images = storage.NewLocalRegistryStorage(nil)
	assert.False(t, s.IsRemoteImagesStorage())

	s.Images = &storage.RepoRegistryStorage{RepoAddress: "registry.example/project"}
	assert.True(t, s.IsRemoteImagesStorage())
}

func TestAI_CustomTagsStorage_PrefersFinalOverImages(t *testing.T) {
	images := storage.NewLocalRegistryStorage(nil)
	final := &storage.RepoRegistryStorage{RepoAddress: "registry.example/final"}

	s := Storages{Images: images}
	assert.Same(t, images, s.CustomTags())

	s.Final = final
	assert.Same(t, final, s.CustomTags())
}
