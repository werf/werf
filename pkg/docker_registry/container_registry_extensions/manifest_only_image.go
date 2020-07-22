package container_registry_extensions

import (
	"fmt"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type manifestOnlyImage struct {
	CreatedAt time.Time
	Labels    map[string]string
}

func NewManifestOnlyImage(labels map[string]string) v1.Image {
	if img, err := partial.UncompressedToImage(manifestOnlyImage{
		CreatedAt: time.Now(),
		Labels:    labels,
	}); err != nil {
		panic(fmt.Sprintf("unable to create new ManifestOnlyImage: %s", err))
	} else {
		return img
	}
}

// MediaType implements partial.UncompressedImageCore.
func (i manifestOnlyImage) MediaType() (types.MediaType, error) {
	return types.DockerManifestSchema2, nil
}

// RawConfigFile implements partial.UncompressedImageCore.
func (i manifestOnlyImage) RawConfigFile() ([]byte, error) {
	return partial.RawConfigFile(i)
}

// ConfigFile implements v1.Image.
func (i manifestOnlyImage) ConfigFile() (*v1.ConfigFile, error) {
	return &v1.ConfigFile{
		Created: v1.Time{i.CreatedAt},
		Config: v1.Config{
			Labels: i.Labels,
		},
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: []v1.Hash{EmptyUncompressedLayer.diffID},
		},
	}, nil
}

func (i manifestOnlyImage) LayerByDiffID(h v1.Hash) (partial.UncompressedLayer, error) {
	return EmptyUncompressedLayer, nil
}
