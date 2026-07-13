package build

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
	imagePkg "github.com/werf/werf/v2/pkg/image"
)

func newStageImageWithDesc(t *testing.T, desc *imagePkg.StageDesc) *stage.StageImage {
	t.Helper()
	img := container_backend.NewLegacyStageImage("prev-stage", nil, "linux/amd64")
	if desc != nil {
		img.SetStageDesc(desc)
	}
	return stage.NewStageImage(nil, "", img)
}

func TestAI_resolvePrevStageBaseImage_BuildkitPrefersRepoDigest(t *testing.T) {
	prev := newStageImageWithDesc(t, &imagePkg.StageDesc{
		Info: &imagePkg.Info{
			Name:       "registry.example/proj:digest-123",
			ID:         "sha256:cfabee873c3dbeda9b1b1691c4b701940dfe6f7b605dcf9244cb62185bda14f0",
			RepoDigest: "registry.example/proj@sha256:bbeb05a26fcc434aaa53214824fb71d88087db4f17e17332ed7742d428d23893",
		},
	})

	got := resolvePrevStageBaseImage(prev, true)
	assert.Equal(t, "registry.example/proj@sha256:bbeb05a26fcc434aaa53214824fb71d88087db4f17e17332ed7742d428d23893", got)
}

func TestAI_resolvePrevStageBaseImage_BuildkitFallsBackToName(t *testing.T) {
	prev := newStageImageWithDesc(t, &imagePkg.StageDesc{
		Info: &imagePkg.Info{
			Name: "registry.example/proj:digest-123",
			ID:   "sha256:cfabee873c3dbeda9b1b1691c4b701940dfe6f7b605dcf9244cb62185bda14f0",
		},
	})

	got := resolvePrevStageBaseImage(prev, true)
	assert.Equal(t, "registry.example/proj:digest-123", got)
}

func TestAI_resolvePrevStageBaseImage_DockerBackendReturnsConfigDigest(t *testing.T) {
	prev := newStageImageWithDesc(t, &imagePkg.StageDesc{
		Info: &imagePkg.Info{
			Name:       "registry.example/proj:digest-123",
			ID:         "sha256:cfabee873c3dbeda9b1b1691c4b701940dfe6f7b605dcf9244cb62185bda14f0",
			RepoDigest: "registry.example/proj@sha256:bbeb05a26fcc434aaa53214824fb71d88087db4f17e17332ed7742d428d23893",
		},
	})

	got := resolvePrevStageBaseImage(prev, false)
	assert.Equal(t, "sha256:cfabee873c3dbeda9b1b1691c4b701940dfe6f7b605dcf9244cb62185bda14f0", got)
}

func TestAI_resolvePrevStageBaseImage_NilOrEmptyReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", resolvePrevStageBaseImage(nil, true))
	assert.Equal(t, "", resolvePrevStageBaseImage(nil, false))

	prev := newStageImageWithDesc(t, nil)
	assert.Equal(t, "prev-stage", resolvePrevStageBaseImage(prev, true))
	assert.Equal(t, "prev-stage", resolvePrevStageBaseImage(prev, false))
}
