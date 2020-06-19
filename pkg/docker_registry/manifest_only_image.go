package docker_registry

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type manifestOnlyImage struct {
	Labels map[string]string
}

func NewManifestOnlyImage(labels map[string]string) v1.Image {
	if img, err := partial.UncompressedToImage(manifestOnlyImage{Labels: labels}); err != nil {
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
		Config: v1.Config{
			Labels: i.Labels,
		},
		RootFS: v1.RootFS{
			// Some clients check this.
			Type: "layers",
		},
	}, nil
}

func (i manifestOnlyImage) LayerByDiffID(h v1.Hash) (partial.UncompressedLayer, error) {
	return nil, fmt.Errorf("LayerByDiffID(%s): empty image", h)
}
