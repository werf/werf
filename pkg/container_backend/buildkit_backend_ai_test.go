package container_backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_BuildkitBackend_RepoNotSetError(t *testing.T) {
	backend := NewBuildkitBackend("tcp://localhost:1234", BuildkitBackendOptions{})

	_, err := backend.getStagesStorageRepo()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--repo is required when using buildkit backend")

	backend.SetStagesStorage("registry.example.com/project", nil)
	repo, err := backend.getStagesStorageRepo()
	require.NoError(t, err)
	assert.Equal(t, "registry.example.com/project", repo)
}

func TestAI_AsBuildkitBackend_UnwrapsPerfCheck(t *testing.T) {
	backend := NewBuildkitBackend("tcp://localhost:1234", BuildkitBackendOptions{})

	unwrapped, ok := AsBuildkitBackend(NewPerfCheckContainerBackend(backend))
	require.True(t, ok)
	assert.Same(t, backend, unwrapped)

	_, ok = AsBuildkitBackend(NewDockerServerBackend(nil))
	assert.False(t, ok)
}
