package deno

import (
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/nelm/pkg/ts/denolock"
)

func TestEmbedPlatformListsInSync(t *testing.T) {
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
	slices.Sort(embedPlatforms)

	taskfileSrc, err := os.ReadFile(filepath.Join("..", "..", "Taskfile.dist.yaml"))
	require.NoError(t, err)

	taskfileRe := regexp.MustCompile(`- build:dist:([a-z0-9]+):([a-z0-9]+)`)

	releaseSeen := make(map[string]struct{})
	for _, m := range taskfileRe.FindAllStringSubmatch(string(taskfileSrc), -1) {
		releaseSeen[m[1]+"/"+m[2]] = struct{}{}
	}

	require.NotEmpty(t, releaseSeen)

	releasePlatforms := slices.Sorted(maps.Keys(releaseSeen))

	lock, err := denolock.Read()
	require.NoError(t, err)

	pinnedPlatforms := slices.Sorted(maps.Keys(lock.Platforms))

	assert.Equal(t, releasePlatforms, embedPlatforms,
		"pkg/deno/embed_*.go files out of sync with build:dist:all release targets in Taskfile.dist.yaml")
	assert.Equal(t, pinnedPlatforms, embedPlatforms,
		"pkg/deno/embed_*.go files out of sync with the platforms nelm pins in pkg/ts/denolock")
}
