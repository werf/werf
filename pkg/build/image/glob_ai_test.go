package image

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_ExtendedGlob(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a/b"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "top.txt"), nil, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a/mid.txt"), nil, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a/b/deep.txt"), nil, 0o644))

	matches, err := extendedGlob(filepath.Join(dir, "*.txt"))
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "top.txt")}, matches)

	matches, err = extendedGlob(filepath.Join(dir, "**", "*.txt"))
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "a/b/deep.txt"),
		filepath.Join(dir, "a/mid.txt"),
		filepath.Join(dir, "top.txt"),
	}, matches)

	matches, err = extendedGlob(filepath.Join(dir, "missing.txt"))
	require.NoError(t, err)
	assert.Empty(t, matches)

	matches, err = extendedGlob(filepath.Join(dir, "a"))
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "a")}, matches)
}
