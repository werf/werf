package deno

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_EmbedPlatformListsInSync(t *testing.T) {
	embedFiles, err := filepath.Glob("embed_*.go")
	require.NoError(t, err)

	fileRe := regexp.MustCompile(`^embed_([a-z0-9]+)_([a-z0-9]+)\.go$`)

	var embedPlatforms []string
	for _, f := range embedFiles {
		m := fileRe.FindStringSubmatch(filepath.Base(f))
		require.NotNil(t, m, "unexpected embed file name: %s", f)
		embedPlatforms = append(embedPlatforms, m[1]+"/"+m[2])
	}

	require.NotEmpty(t, embedPlatforms)
	sort.Strings(embedPlatforms)

	taskfileSrc, err := os.ReadFile(filepath.Join("..", "..", "Taskfile.dist.yaml"))
	require.NoError(t, err)

	taskfileRe := regexp.MustCompile(`- build:dist:([a-z0-9]+):([a-z0-9]+)`)

	var releasePlatforms []string
	for _, m := range taskfileRe.FindAllStringSubmatch(string(taskfileSrc), -1) {
		releasePlatforms = append(releasePlatforms, m[1]+"/"+m[2])
	}

	require.NotEmpty(t, releasePlatforms)
	sort.Strings(releasePlatforms)

	assert.Equal(t, releasePlatforms, embedPlatforms, "pkg/deno/embed_*.go files out of sync with build:dist:all release targets in Taskfile.dist.yaml")
}
