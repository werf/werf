//go:build embedwerfdeno

package deno

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_EmbeddedDenoDataAvailable(t *testing.T) {
	compressed, expectedSha256, ok := EmbeddedDenoData()

	require.True(t, ok)
	assert.NotEmpty(t, compressed)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{64}$`), expectedSha256)
}

func TestAI_EmbeddedDenoDataMatchesSha256(t *testing.T) {
	compressed, expectedSha256, ok := EmbeddedDenoData()
	require.True(t, ok)

	gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)

	defer gzReader.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, gzReader)
	require.NoError(t, err)

	assert.Equal(t, expectedSha256, hex.EncodeToString(hasher.Sum(nil)))
}
