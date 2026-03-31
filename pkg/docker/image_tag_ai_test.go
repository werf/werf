//go:build ai_tests

package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Tag_Success(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	sourceImage := "alpine:latest"
	targetImage := "test-werf-tag-ai:latest"

	err = CliPull(ctx, sourceImage)
	require.NoError(t, err, "Failed to pull source image")

	defer func() {
		_ = CliRmi(ctx, sourceImage, targetImage)
	}()

	err = CliTag(ctx, sourceImage, targetImage)
	require.NoError(t, err, "CliTag should succeed")

	inspect, err := ImageInspect(ctx, targetImage)
	require.NoError(t, err, "Target image should exist after tagging")
	assert.NotNil(t, inspect, "ImageInspect should return data for tagged image")
}

func TestAI_Tag_InvalidSource(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	err = CliTag(ctx, "nonexistent-image-12345:latest", "test:latest")
	require.Error(t, err, "CliTag should fail with nonexistent source image")
}

func TestAI_Tag_InsufficientArgs(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	err = CliTag(ctx, "only-one-arg")
	require.Error(t, err, "CliTag should fail with insufficient arguments")
}
