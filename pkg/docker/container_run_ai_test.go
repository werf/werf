//go:build ai_tests

package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Run_Success(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	err = CliRun(ctx, "--rm", "alpine:latest", "echo", "hello")
	require.NoError(t, err)
}

func TestAI_Run_RecordedOutput(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	output, err := CliRun_RecordedOutput(ctx, "--rm", "alpine:latest", "echo", "hello", "world")
	require.NoError(t, err)
	assert.Contains(t, output, "hello world")
}

func TestAI_Run_LiveOutput(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	err = CliRun_LiveOutput(ctx, "--rm", "alpine:latest", "echo", "hello")
	require.NoError(t, err)
}

func TestAI_Run_NonZeroExit(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	_, err = CliRun_RecordedOutput(ctx, "--rm", "alpine:latest", "sh", "-c", "exit 42")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "42")
}

func TestAI_Run_WithVolume(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	_, err = CliRun_RecordedOutput(ctx, "--rm", "--volume", "/tmp:/mnt", "alpine:latest", "ls", "/mnt")
	require.NoError(t, err)
}
