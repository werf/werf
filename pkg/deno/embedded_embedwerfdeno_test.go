//go:build embedwerfdeno

package deno

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/nelm/pkg/ts/denolock"
)

func TestEmbeddedDenoDataAvailable(t *testing.T) {
	compressed, ok := EmbeddedDenoData()

	require.True(t, ok)
	assert.NotEmpty(t, compressed)
}

// The blob werf ships must be the Deno release nelm pins, since that is the only digest nelm will
// accept when extracting it.
func TestEmbeddedDenoDataMatchesPinnedRelease(t *testing.T) {
	compressed, ok := EmbeddedDenoData()
	require.True(t, ok)

	pinned, err := denolock.Get(runtime.GOOS, runtime.GOARCH)
	require.NoError(t, err)

	gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)

	defer gzReader.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, gzReader)
	require.NoError(t, err)

	assert.Equal(t, pinned.BinarySHA256, hex.EncodeToString(hasher.Sum(nil)))
}
