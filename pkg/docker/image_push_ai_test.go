//go:build ai_tests

package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Push_InvalidRegistry(t *testing.T) {
	ctx := context.Background()

	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available:", err)
	}

	err = doCliPush(ctx, "localhost:9999/nonexistent-image:test")
	require.Error(t, err)
}

func TestAI_Push_AuthHeaderConstructed(t *testing.T) {
	ctx := context.Background()

	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available:", err)
	}

	auth, err := getRegistryAuth(ctx, "docker.io/library/alpine:latest")
	require.NoError(t, err)
	assert.NotEmpty(t, auth)
}
