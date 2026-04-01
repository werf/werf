//go:build ai_tests

package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Create_Success(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		err = CliPull(ctx, "alpine:latest")
		require.NoError(t, err)
	}

	containerName := "testcontainer-ai-create"

	defer func() {
		_ = ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
	}()

	err = CliCreate(ctx, "--name="+containerName, "--volume=/tmp:/tmp", "alpine:latest")
	require.NoError(t, err)

	inspect, err := ContainerInspect(ctx, containerName)
	require.NoError(t, err)
	assert.Equal(t, "/"+containerName, inspect.Name)
	assert.NotEmpty(t, inspect.Mounts)

	foundVolume := false
	for _, mount := range inspect.Mounts {
		if mount.Type == "bind" && mount.Source == "/tmp" && mount.Destination == "/tmp" {
			foundVolume = true
			break
		}
	}
	assert.True(t, foundVolume, "Expected /tmp:/tmp volume mount not found")
}

func TestAI_Create_InvalidImage(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	err := CliCreate(ctx, "--name=testcontainer-ai-invalid", "nonexistent-image:99999")
	assert.Error(t, err)
}
