package buildkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_HostFromEnv_WerfEnvWinsOverBare(t *testing.T) {
	t.Setenv("WERF_BUILDKIT_HOST", "tcp://werf-host:1234")
	t.Setenv("BUILDKIT_HOST", "tcp://bare-host:1234")
	assert.Equal(t, "tcp://werf-host:1234", HostFromEnv())
}

func TestAI_HostFromEnv_FallsBackToBare(t *testing.T) {
	t.Setenv("WERF_BUILDKIT_HOST", "")
	t.Setenv("BUILDKIT_HOST", "unix:///run/buildkit/buildkitd.sock")
	assert.Equal(t, "unix:///run/buildkit/buildkitd.sock", HostFromEnv())
}

func TestAI_GetHost_ErrorWhenUnset(t *testing.T) {
	t.Setenv("WERF_BUILDKIT_HOST", "")
	t.Setenv("BUILDKIT_HOST", "")

	_, err := GetHost()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WERF_BUILDKIT_HOST")
	assert.Contains(t, err.Error(), "BUILDKIT_HOST")
	assert.Contains(t, err.Error(), "docker run")
}
