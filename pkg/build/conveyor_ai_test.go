package build

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
)

func TestAI_Conveyor_StageImageCacheSeparatesPlatforms(t *testing.T) {
	conveyor := &Conveyor{
		stageImages:      make(map[string]*stage.StageImage),
		serviceRWMutex:   map[string]*sync.RWMutex{},
		stageDigestMutex: map[string]*sync.Mutex{},
	}

	amd64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/amd64"))
	arm64Image := stage.NewStageImage(nil, "", container_backend.NewLegacyStageImage(nil, "alpine:3.22.3", nil, "linux/arm64"))

	conveyor.SetStageImage(amd64Image)
	conveyor.SetStageImage(arm64Image)

	require.Same(t, amd64Image, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64"))
	require.Same(t, arm64Image, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64"))
	assert.NotSame(t, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64"), conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64"))
	assert.Nil(t, conveyor.GetStageImage("alpine:3.22.3"))

	conveyor.UnsetStageImageByPlatform("alpine:3.22.3", "linux/amd64")
	assert.Nil(t, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64"))
	assert.Same(t, arm64Image, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64"))
	assert.Same(t, arm64Image, conveyor.GetStageImage("alpine:3.22.3"))

	conveyor.UnsetStageImage("alpine:3.22.3")
	assert.Nil(t, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/amd64"))
	assert.Nil(t, conveyor.GetStageImageByPlatform("alpine:3.22.3", "linux/arm64"))
}
