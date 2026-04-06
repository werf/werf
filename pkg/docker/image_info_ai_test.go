package docker

import (
	"testing"

	dockerImage "github.com/docker/docker/api/types/image"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/werf/werf/v2/pkg/image"
)

func TestAI_NewInfoFromInspect_LabelPresent(t *testing.T) {
	inspect := &dockerImage.InspectResponse{
		ID:      "sha256:abc123",
		Created: "2024-01-01T00:00:00Z",
		Config: &dockerspec.DockerOCIImageConfig{
			DockerOCIImageConfigExt: dockerspec.DockerOCIImageConfigExt{},
			ImageConfig: ocispec.ImageConfig{
				Labels: map[string]string{
					image.WerfBaseImageIDLabel: "sha256:parent123",
				},
			},
		},
	}

	info := NewInfoFromInspect("myimage:tag", inspect)
	assert.Equal(t, "sha256:parent123", info.ParentID)
	assert.Equal(t, "sha256:abc123", info.ID)
}

func TestAI_NewInfoFromInspect_LabelAbsent(t *testing.T) {
	inspect := &dockerImage.InspectResponse{
		ID:      "sha256:abc123",
		Created: "2024-01-01T00:00:00Z",
		Config: &dockerspec.DockerOCIImageConfig{
			ImageConfig: ocispec.ImageConfig{
				Labels: map[string]string{
					"some.other.label": "value",
				},
			},
		},
	}

	info := NewInfoFromInspect("myimage:tag", inspect)
	assert.Empty(t, info.ParentID)
}

func TestAI_NewInfoFromInspect_NilLabels(t *testing.T) {
	inspect := &dockerImage.InspectResponse{
		ID:      "sha256:abc123",
		Created: "2024-01-01T00:00:00Z",
		Config:  &dockerspec.DockerOCIImageConfig{},
	}

	info := NewInfoFromInspect("myimage:tag", inspect)
	assert.Empty(t, info.ParentID)
}
