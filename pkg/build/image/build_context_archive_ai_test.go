package image

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/container_backend"
)

func newBuildContextArchiveAI(t *testing.T, dirName string) *BuildContextArchive {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, dirName), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, dirName, "x"), []byte("data"), 0o644))

	return &BuildContextArchive{path: filepath.Join(root, "context.tar"), extractionDir: root}
}

func TestAI_CalculateGlobsChecksumMatchedPaths(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name  string
		opts  container_backend.CalculateGlobsChecksumOptions
		equal bool
	}{
		{"contents only", container_backend.CalculateGlobsChecksumOptions{}, true},
		{"matched paths included", container_backend.CalculateGlobsChecksumOptions{IncludeMatchedPaths: true}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checksumAAA, err := newBuildContextArchiveAI(t, "aaa").CalculateGlobsChecksum(ctx, []string{"./*"}, tt.opts)
			require.NoError(t, err)

			checksumCCC, err := newBuildContextArchiveAI(t, "ccc").CalculateGlobsChecksum(ctx, []string{"./*"}, tt.opts)
			require.NoError(t, err)

			if tt.equal {
				require.Equal(t, checksumAAA, checksumCCC)
			} else {
				require.NotEqual(t, checksumAAA, checksumCCC)
			}
		})
	}
}
