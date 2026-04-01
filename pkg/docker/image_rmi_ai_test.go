//go:build ai_tests

package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_Rmi_Success(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	require.NoError(t, err)

	err = CliPull(ctx, "alpine:latest")
	require.NoError(t, err)

	err = CliTag(ctx, "alpine:latest", "test-werf-rmi-ai:latest")
	require.NoError(t, err)

	images, err := Images(ctx, types.ImageListOptions{})
	require.NoError(t, err)
	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == "test-werf-rmi-ai:latest" {
				found = true
				break
			}
		}
	}
	require.True(t, found, "test-werf-rmi-ai:latest tag should exist")

	err = CliRmi(ctx, "test-werf-rmi-ai:latest")
	require.NoError(t, err)

	images, err = Images(ctx, types.ImageListOptions{})
	require.NoError(t, err)
	tagGone := true
	alpineExists := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == "test-werf-rmi-ai:latest" {
				tagGone = false
			}
			if tag == "alpine:latest" {
				alpineExists = true
			}
		}
	}
	assert.True(t, tagGone, "test-werf-rmi-ai:latest tag should be removed")
	assert.True(t, alpineExists, "alpine:latest should still exist")
}

func TestAI_Rmi_Force(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	require.NoError(t, err)

	err = CliPull(ctx, "alpine:latest")
	require.NoError(t, err)

	err = CliTag(ctx, "alpine:latest", "test-werf-rmi-force-ai:latest")
	require.NoError(t, err)

	err = CliRmi(ctx, "--force", "test-werf-rmi-force-ai:latest")
	require.NoError(t, err)

	images, err := Images(ctx, types.ImageListOptions{})
	require.NoError(t, err)
	tagGone := true
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == "test-werf-rmi-force-ai:latest" {
				tagGone = false
				break
			}
		}
	}
	assert.True(t, tagGone, "test-werf-rmi-force-ai:latest tag should be removed")
}

func TestAI_Rmi_NotFound(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx, InitOptions{})
	require.NoError(t, err)

	err = CliRmi(ctx, "nonexistent-image-12345:latest")
	require.Error(t, err, "should return error for non-existent image")
}
