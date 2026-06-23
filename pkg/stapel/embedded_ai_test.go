package stapel

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_embeddedImageForPlatform(t *testing.T) {
	tests := []struct {
		name           string
		targetPlatform string
		wantOK         bool
	}{
		{name: "linux/amd64", targetPlatform: "linux/amd64", wantOK: true},
		{name: "linux/arm64", targetPlatform: "linux/arm64", wantOK: true},
		{name: "windows/amd64", targetPlatform: "windows/amd64", wantOK: false},
		{name: "linux/arm/v7", targetPlatform: "linux/arm/v7", wantOK: false},
		{name: "garbage", targetPlatform: "not-a-platform/x/y/z", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, ok := embeddedImageForPlatform(tt.targetPlatform)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.NotEmpty(t, img.gzipData)
				assert.NotEmpty(t, img.expectedSha256)
			}
		})
	}
}

func TestAI_normalizeEmbeddedPlatform(t *testing.T) {
	tests := []struct {
		name           string
		targetPlatform string
		want           string
		wantErr        bool
	}{
		{name: "amd64", targetPlatform: "linux/amd64", want: "linux/amd64"},
		{name: "x86_64 alias", targetPlatform: "linux/x86_64", want: "linux/amd64"},
		{name: "aarch64 alias", targetPlatform: "linux/aarch64", want: "linux/arm64"},
		{name: "invalid", targetPlatform: "///", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEmbeddedPlatform(tt.targetPlatform)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAI_decompressAndVerify_valid(t *testing.T) {
	payload := []byte("stapel tar payload")

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	sum := sha256.Sum256(payload)
	expected := hex.EncodeToString(sum[:])

	got, err := decompressAndVerify(buf.Bytes(), expected)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
}

func TestAI_decompressAndVerify_mismatch(t *testing.T) {
	payload := []byte("stapel tar payload")

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	_, err = decompressAndVerify(buf.Bytes(), "0000000000000000000000000000000000000000000000000000000000000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integrity check failed")
}

func TestAI_decompressAndVerify_invalidGzip(t *testing.T) {
	_, err := decompressAndVerify([]byte("not gzip data"), "deadbeef")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gzip")
}

func TestAI_isDefaultImageRef(t *testing.T) {
	t.Setenv("WERF_STAPEL_IMAGE_NAME", "")
	t.Setenv("WERF_STAPEL_IMAGE_VERSION", "")
	assert.True(t, isDefaultImageRef())

	t.Setenv("WERF_STAPEL_IMAGE_NAME", "custom.registry/stapel")
	assert.False(t, isDefaultImageRef())

	t.Setenv("WERF_STAPEL_IMAGE_NAME", "")
	t.Setenv("WERF_STAPEL_IMAGE_VERSION", "9.9.9")
	assert.False(t, isDefaultImageRef())
}

func TestAI_embeddedArtifactsMatchSha256(t *testing.T) {
	for _, platform := range []string{"linux/amd64", "linux/arm64"} {
		t.Run(platform, func(t *testing.T) {
			img, ok := embeddedImageForPlatform(platform)
			require.True(t, ok)

			_, err := decompressAndVerify(img.gzipData, img.expectedSha256)
			require.NoError(t, err)
		})
	}
}
