package buildkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestAI_ResolveHost_EnvWinsWithoutDocker(t *testing.T) {
	t.Setenv("WERF_BUILDKIT_HOST", "tcp://werf-host:1234")

	host, err := ResolveHost(t.Context(), ResolveHostOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "tcp://werf-host:1234", host)
}

func TestAI_ResolveHost_ErrorMentionsEnvVarsWhenDockerUnavailable(t *testing.T) {
	t.Setenv("WERF_BUILDKIT_HOST", "")
	t.Setenv("BUILDKIT_HOST", "")
	t.Setenv("DOCKER_HOST", "unix:///nonexistent/docker.sock")

	_, err := ResolveHost(t.Context(), ResolveHostOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "WERF_BUILDKIT_HOST")
	assert.Contains(t, err.Error(), "BUILDKIT_HOST")
}

func TestAI_MakeLocalBuildkitdConfig_MapsFlagsToRegistryOptions(t *testing.T) {
	config, err := makeLocalBuildkitdConfig(ResolveHostOptions{
		InsecureRegistryAddresses:      []string{"192.168.1.66:5556/werf-stapel-test"},
		SkipTLSVerifyRegistryAddresses: []string{"192.168.1.66:5556/werf-stapel-test", "my.registry.example/proj"},
	})
	assert.NoError(t, err)
	assert.Equal(t, `[registry."192.168.1.66:5556"]
  http = true
  insecure = true
[registry."my.registry.example"]
  insecure = true
`, config)
}

func TestAI_MakeLocalBuildkitdConfig_EmptyWithoutFlags(t *testing.T) {
	config, err := makeLocalBuildkitdConfig(ResolveHostOptions{})
	assert.NoError(t, err)
	assert.Empty(t, config)
}
