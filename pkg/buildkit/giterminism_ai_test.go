package buildkit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Giterminism requires all build inputs to come from materialized git commit trees
// delivered via llb.Local; llb.Git sources must never be used.
func TestAI_Giterminism_NoLLBGitSources(t *testing.T) {
	dirs := []string{".", "../container_backend"}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, name))
			require.NoError(t, err)
			assert.NotContains(t, string(data), "llb.Git(", "llb.Git source is forbidden (giterminism): %s", filepath.Join(dir, name))
		}
	}
}
