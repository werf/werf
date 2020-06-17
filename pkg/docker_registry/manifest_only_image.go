package docker_registry

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type ManifestOnlyImage struct {
	Metadata map[string]string
}

func NewManifestOnlyImage(metadata map[string]string) v1.Image {
	if img, err := partial.UncompressedToImage(ManifestOnlyImage{Metadata: metadata}); err != nil {
		panic(fmt.Sprintf("unable to create new ManifestOnlyImage: %s", err))
	} else {
		return img
	}
}

// MediaType implements partial.UncompressedImageCore.
func (i ManifestOnlyImage) MediaType() (types.MediaType, error) {
	return types.DockerManifestSchema2, nil
}

// RawConfigFile implements partial.UncompressedImageCore.
func (i ManifestOnlyImage) RawConfigFile() ([]byte, error) {
	return partial.RawConfigFile(i)
}

// ConfigFile implements v1.Image.
func (i ManifestOnlyImage) ConfigFile() (*v1.ConfigFile, error) {
	return &v1.ConfigFile{
		Config: v1.Config{
			Labels: i.Metadata,
		},
		RootFS: v1.RootFS{
			// Some clients check this.
			Type: "layers",
		},
	}, nil
}

func (i ManifestOnlyImage) LayerByDiffID(h v1.Hash) (partial.UncompressedLayer, error) {
	return nil, fmt.Errorf("LayerByDiffID(%s): empty image", h)
}
