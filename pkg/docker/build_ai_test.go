//go:build ai_tests

package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createBuildContext(t *testing.T, files map[string]string) io.ReadCloser {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	return io.NopCloser(&buf)
}

func TestAI_Build_SimpleDockerfile(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	tag := "werf-test-build-simple:latest"
	defer func() { _ = CliRmi(ctx, tag) }()

	metadataFile := filepath.Join(t.TempDir(), "metadata.json")
	rc := createBuildContext(t, map[string]string{"Dockerfile": "FROM alpine:latest\nRUN echo hello\n"})

	err := CliBuild_LiveOutputWithCustomIn(ctx, rc,
		"--file", "Dockerfile",
		"--tag", tag,
		"--metadata-file", metadataFile,
		"-",
	)
	require.NoError(t, err)

	data, err := os.ReadFile(metadataFile)
	require.NoError(t, err)

	var meta map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &meta))
	digest, ok := meta["containerimage.digest"].(string)
	require.True(t, ok, "metadata should contain containerimage.digest")
	assert.Contains(t, digest, "sha256:")

	exists, err := ImageExist(ctx, tag)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAI_Build_WithBuildArgs(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	tag := "werf-test-build-args:latest"
	defer func() { _ = CliRmi(ctx, tag) }()

	metadataFile := filepath.Join(t.TempDir(), "metadata.json")
	rc := createBuildContext(t, map[string]string{
		"Dockerfile": "FROM alpine:latest\nARG MY_VAR=default\nRUN echo $MY_VAR > /test.txt\n",
	})

	err := CliBuild_LiveOutputWithCustomIn(ctx, rc,
		"--file", "Dockerfile",
		"--tag", tag,
		"--build-arg", "MY_VAR=testvalue",
		"--metadata-file", metadataFile,
		"-",
	)
	require.NoError(t, err)

	exists, err := ImageExist(ctx, tag)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAI_Build_WithTarget(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	tag := "werf-test-build-target:latest"
	defer func() { _ = CliRmi(ctx, tag) }()

	metadataFile := filepath.Join(t.TempDir(), "metadata.json")
	rc := createBuildContext(t, map[string]string{
		"Dockerfile": "FROM alpine:latest AS stage1\nRUN echo stage1\n\nFROM alpine:latest AS stage2\nRUN echo stage2\n",
	})

	err := CliBuild_LiveOutputWithCustomIn(ctx, rc,
		"--file", "Dockerfile",
		"--tag", tag,
		"--target", "stage1",
		"--metadata-file", metadataFile,
		"-",
	)
	require.NoError(t, err)

	exists, err := ImageExist(ctx, tag)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAI_Build_WithPlatform(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	tag := "werf-test-build-platform:latest"
	defer func() { _ = CliRmi(ctx, tag) }()

	metadataFile := filepath.Join(t.TempDir(), "metadata.json")
	rc := createBuildContext(t, map[string]string{
		"Dockerfile": "FROM alpine:latest\nRUN echo hello\n",
	})

	err := CliBuild_LiveOutputWithCustomIn(ctx, rc,
		"--file", "Dockerfile",
		"--tag", tag,
		"--platform", "linux/amd64",
		"--metadata-file", metadataFile,
		"-",
	)
	require.NoError(t, err)

	exists, err := ImageExist(ctx, tag)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAI_Build_InvalidDockerfile(t *testing.T) {
	ctx := context.Background()
	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	metadataFile := filepath.Join(t.TempDir(), "metadata.json")
	rc := createBuildContext(t, map[string]string{
		"Dockerfile": "INVALID_INSTRUCTION something\n",
	})

	err := CliBuild_LiveOutputWithCustomIn(ctx, rc,
		"--file", "Dockerfile",
		"--tag", "werf-test-build-invalid:latest",
		"--metadata-file", metadataFile,
		"-",
	)
	require.Error(t, err)
}
