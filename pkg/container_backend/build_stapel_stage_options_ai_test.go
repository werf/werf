package container_backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAI_BuildStapelStageOptions_BuildTimeEnvsSeparateFromEnvs(t *testing.T) {
	opts := &BuildStapelStageOptions{}
	opts.AddEnvs(map[string]string{"PERSISTENT": "1"})
	opts.AddBuildTimeEnvs(map[string]string{"SSH_AUTH_SOCK": "/tmp/ssh.sock"})

	assert.Equal(t, map[string]string{"PERSISTENT": "1"}, opts.Envs)
	assert.Equal(t, map[string]string{"SSH_AUTH_SOCK": "/tmp/ssh.sock"}, opts.BuildTimeEnvs)

	mergedEnvs := make(map[string]string, len(opts.Envs)+len(opts.BuildTimeEnvs))
	for k, v := range opts.Envs {
		mergedEnvs[k] = v
	}
	for k, v := range opts.BuildTimeEnvs {
		mergedEnvs[k] = v
	}

	assert.Equal(t, map[string]string{"PERSISTENT": "1", "SSH_AUTH_SOCK": "/tmp/ssh.sock"}, mergedEnvs)
}
