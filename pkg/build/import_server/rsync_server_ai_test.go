package import_server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_buildRsyncdConf(t *testing.T) {
	conf := buildRsyncdConf("873", "werf-deadbeef")

	require.Contains(t, conf, "exclude = /proc /sys /dev /run /.werf")
	assert.Contains(t, conf, "port = 873")
	assert.Contains(t, conf, "auth users = werf-deadbeef")

	for _, dir := range systemExcludeDirs {
		assert.True(t, strings.HasPrefix(dir, "/"), "system exclude dir %q must be anchored to module root", dir)
	}
}
