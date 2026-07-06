package container_backend

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/werf/logboek"
)

func TestAI_PrepareBuildTimeEnvExports(t *testing.T) {
	ctx := logboek.NewContext(context.Background(), logboek.NewLogger(io.Discard, io.Discard))

	c := &LegacyStageImageContainer{buildTimeEnv: make(map[string]string)}
	c.AddBuildTimeEnv(map[string]string{
		"WERF_COMMIT_HASH":       "abc123",
		"WERF_COMMIT_TIME_HUMAN": "Mon Jan 2 15:04:05 2006",
		"WITH_QUOTE":             "it's",
	})

	exports := c.prepareBuildTimeEnvExports(ctx)

	columnsExport := fmt.Sprintf("export COLUMNS='%d'", logboek.Context(ctx).Streams().ContentWidth())

	assert.Contains(t, exports, columnsExport)
	assert.Contains(t, exports, "export WERF_COMMIT_HASH='abc123'")
	assert.Contains(t, exports, "export WERF_COMMIT_TIME_HUMAN='Mon Jan 2 15:04:05 2006'")
	assert.Contains(t, exports, `export WITH_QUOTE='it'\''s'`)
}
