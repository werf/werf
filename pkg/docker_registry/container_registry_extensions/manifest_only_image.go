package container_registry_extensions

import (
	"fmt"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type manifestOnlyImage struct {
	CreatedAt time.Time
	Labels    map[string]string
}

func NewManifestOnlyImage(labels map[string]string) v1.Image {
	img, err := newManifestOnlyImage(labels)
	if err != nil {
		panic(fmt.Sprintf("unable to create new ManifestOnlyImage: %s", err))
	}

	return img
}

func newManifestOnlyImage(labels map[string]string) (v1.Image, error) {
	t := time.Now()

	img, err := partial.UncompressedToImage(manifestOnlyImage{
		CreatedAt: t,
		Labels:    labels,
	})
	if err != nil {
		return nil, err
	}

	img, err = mutate.CreatedAt(img, v1.Time{Time: t})
	if err != nil {
		return nil, err
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}

	cfg.History = make([]v1.History, len(layers))
	img, err = mutate.ConfigFile(img, cfg)
	if err != nil {
		return nil, err
	}

	return img, nil
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
		Created: v1.Time{Time: i.CreatedAt},
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
