//go:build ai_tests

package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Pull_Success(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available")
	}

	imageRef := "alpine:3.18"

	cleanupImage := func() {
		_ = CliRmi(ctx, imageRef)
	}
	cleanupImage()
	defer cleanupImage()

	err = CliPull(ctx, imageRef)
	require.NoError(t, err)

	_, _, err = apiCli(ctx).ImageInspectWithRaw(ctx, imageRef)
	assert.NoError(t, err, "image should exist after pull")
}

func TestAI_Pull_WithPlatform(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available")
	}

	imageRef := "alpine:3.18"
	platform := "linux/amd64"

	cleanupImage := func() {
		_ = CliRmi(ctx, imageRef)
	}
	cleanupImage()
	defer cleanupImage()

	err = CliPull(ctx, "--platform", platform, imageRef)
	require.NoError(t, err)

	inspect, _, err := apiCli(ctx).ImageInspectWithRaw(ctx, imageRef)
	require.NoError(t, err)
	t.Logf("Image platform: %s/%s", inspect.Os, inspect.Architecture)
}

func TestAI_Pull_NotFound(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available")
	}

	nonExistentImage := "nonexistent.invalid/nope:latest"

	err = CliPull(ctx, nonExistentImage)
	assert.Error(t, err, "should error for non-existent image")
}

func TestAI_Pull_WithRetries_Success(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available")
	}

	imageRef := "alpine:3.18"

	cleanupImage := func() {
		_ = CliRmi(ctx, imageRef)
	}
	cleanupImage()
	defer cleanupImage()

	err = CliPullWithRetries(ctx, imageRef)
	require.NoError(t, err)

	_, _, err = apiCli(ctx).ImageInspectWithRaw(ctx, imageRef)
	assert.NoError(t, err, "image should exist after pull with retries")
}

func TestAI_Pull_WithRetries_WithPlatform(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skip("Docker not available")
	}

	imageRef := "alpine:3.18"
	platform := "linux/amd64"

	cleanupImage := func() {
		_ = CliRmi(ctx, imageRef)
	}
	cleanupImage()
	defer cleanupImage()

	err = CliPullWithRetries(ctx, "--platform", platform, imageRef)
	require.NoError(t, err)

	inspect, _, err := apiCli(ctx).ImageInspectWithRaw(ctx, imageRef)
	require.NoError(t, err)
	t.Logf("Image platform: %s/%s", inspect.Os, inspect.Architecture)
}
