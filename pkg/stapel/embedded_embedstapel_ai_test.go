//go:build embedstapel

package stapel

import (
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
