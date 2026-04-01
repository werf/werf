//go:build ai_tests

package docker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
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

func TestAI_Run_DetachReturnsContainerID(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	containerName := fmt.Sprintf("test-detach-%d", time.Now().UnixNano())
	output, err := CliRun_RecordedOutput(ctx, "--detach", "--rm", fmt.Sprintf("--name=%s", containerName), "alpine:latest", "sleep", "30")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	defer ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
}

func TestAI_Run_DetachContainerIsRunning(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	containerName := fmt.Sprintf("test-running-%d", time.Now().UnixNano())
	_, err = CliRun_RecordedOutput(ctx, "--detach", fmt.Sprintf("--name=%s", containerName), "alpine:latest", "sleep", "30")
	require.NoError(t, err)

	inspection, err := ContainerInspect(ctx, containerName)
	require.NoError(t, err)
	assert.True(t, inspection.State.Running)
	defer ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
}

func TestAI_Run_ExposeFlag(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	err = CliRun(ctx, "--rm", "--expose=8080", "alpine:latest", "echo", "hello")
	require.NoError(t, err)
}

func TestAI_Run_DetachWithRmDoesNotKillContainer(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	containerName := fmt.Sprintf("test-detach-rm-%d", time.Now().UnixNano())
	_, err = CliRun_RecordedOutput(ctx, "--detach", "--rm", fmt.Sprintf("--name=%s", containerName), "alpine:latest", "sleep", "30")
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	inspection, err := ContainerInspect(ctx, containerName)
	require.NoError(t, err)
	assert.True(t, inspection.State.Running)
	defer ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
}

func TestAI_Run_RsyncServerFlagsCombo(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	exist, err := ImageExist(ctx, "alpine:latest")
	require.NoError(t, err)
	if !exist {
		require.NoError(t, CliPull(ctx, "alpine:latest"))
	}

	containerName := fmt.Sprintf("test-rsync-%d", time.Now().UnixNano())
	err = CliRun(ctx,
		"--detach",
		"--rm",
		"--user=0:0",
		"--workdir=/",
		fmt.Sprintf("--name=%s", containerName),
		"--expose=873",
		"--entrypoint=sh",
		"alpine:latest",
		"-c",
		"sleep 10",
	)
	require.NoError(t, err)
	defer ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
}
