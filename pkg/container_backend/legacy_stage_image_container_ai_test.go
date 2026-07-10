package container_backend

import (
	"context"
	"fmt"
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("LegacyStageImageContainer imageRef", func() {
	const (
		tmpUUIDName    = "80324e3c-81f8-43cc-a055-6a95a7928690"
		committedID    = "sha256:deadbeef"
		targetPlatform = "linux/amd64"
	)

	DescribeTable("returns the expected image reference", func(name, platform string, committed bool, expected string) {
		img := NewLegacyStageImage(nil, name, nil, platform)
		if committed {
			img.buildImage = newLegacyBaseImage(committedID, nil)
			img.builtID = committedID
		}

		Expect(img.container.imageRef(img)).To(Equal(expected))
	},
		Entry("committed image with target platform returns committed id, not name", tmpUUIDName, targetPlatform, true, committedID),
		Entry("committed image without target platform returns committed id", tmpUUIDName, "", true, committedID),
		Entry("non-committed image with target platform returns name", "registry.example.com/project:stage", targetPlatform, false, "registry.example.com/project:stage"),
		Entry("non-committed image without target platform returns id from name", "registry.example.com/project:stage", "", false, "registry.example.com/project:stage"),
	)
})
