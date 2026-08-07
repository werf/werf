package manager

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

func TestAI_GenerateStageDescCreationTs_AvoidsNameCollision(t *testing.T) {
	m := &StorageManager{
		ProjectName:   "test-project",
		StagesStorage: storage.NewLocalStagesStorage(nil),
	}
	digest := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	_, baseTs := m.GenerateStageDescCreationTs(digest, image.NewStageDescSet())

	occupied := image.NewStageDescSet()
	occupiedNames := make(map[string]struct{})
	for ts := baseTs; ts <= baseTs+10; ts++ {
		name := m.StagesStorage.ConstructStageImageName(m.ProjectName, digest, ts)
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
