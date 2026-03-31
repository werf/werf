//go:build ai_tests

package docker

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Login_EmptyUsername(t *testing.T) {
	ctx := context.Background()

	err := Login(ctx, "", "password", "registry.example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
}

func TestAI_Login_EmptyPassword(t *testing.T) {
	ctx := context.Background()

	err := Login(ctx, "username", "", "registry.example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestAI_Login_EmptyRegistry(t *testing.T) {
	ctx := context.Background()

	err := Login(ctx, "username", "password", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry address cannot be empty")
}

func TestAI_Login_InvalidCredentials(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" && !fileExists("/var/run/docker.sock") {
		t.Skip("Docker not available, skipping integration test")
	}

	ctx := context.Background()

	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Failed to initialize Docker: %v", err)
	}

	ctx, err := NewContext(ctx)
	require.NoError(t, err)

	err = Login(ctx, "baduser", "badpass", "localhost:5555")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to authenticate")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
