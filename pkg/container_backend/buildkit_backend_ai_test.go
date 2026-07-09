package container_backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/docker_registry"
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

type copyImageRegistryStub struct {
	docker_registry.Interface

	copiedSrc  string
	copiedDest string
}

func (r *copyImageRegistryStub) CopyImage(_ context.Context, src, dest string, _ docker_registry.CopyImageOptions) error {
	r.copiedSrc = src
	r.copiedDest = dest
	return nil
}

func TestAI_BuildkitBackend_TagIsRegistrySide(t *testing.T) {
	registryStub := &copyImageRegistryStub{}
	backend := NewBuildkitBackend("tcp://localhost:1234", BuildkitBackendOptions{})
	backend.SetStagesStorage("registry.example.com/project", registryStub)

	builtID := "registry.example.com/project@sha256:1111111111111111111111111111111111111111111111111111111111111111"
	require.NoError(t, backend.Tag(context.Background(), builtID, "registry.example.com/project:stage-tag", TagOpts{}))
	assert.Equal(t, builtID, registryStub.copiedSrc)
	assert.Equal(t, "registry.example.com/project:stage-tag", registryStub.copiedDest)

	require.NoError(t, backend.Push(context.Background(), "registry.example.com/project:stage-tag", PushOpts{}))
}

func TestAI_BuildkitBackend_SharedClient(t *testing.T) {
	backend := NewBuildkitBackend("tcp://localhost:1234", BuildkitBackendOptions{})

	cl1, err := backend.getClient(context.Background())
	require.NoError(t, err)
	cl2, err := backend.getClient(context.Background())
	require.NoError(t, err)
	assert.Same(t, cl1, cl2)
}

func TestAI_AsBuildkitBackend_UnwrapsPerfCheck(t *testing.T) {
	backend := NewBuildkitBackend("tcp://localhost:1234", BuildkitBackendOptions{})

	unwrapped, ok := AsBuildkitBackend(NewPerfCheckContainerBackend(backend))
	require.True(t, ok)
	assert.Same(t, backend, unwrapped)

	_, ok = AsBuildkitBackend(NewDockerServerBackend(nil))
	assert.False(t, ok)
}
