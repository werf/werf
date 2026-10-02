package container_registry_extensions

import (
	"fmt"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// ManifestFormat selects the media types of the pushed image. Manifest, config and layer media
// types must belong to the same format: containers/image validates layer media types against the
// manifest format and rejects an image that mixes them, which breaks every buildah-based consumer
// of the image and of the stages built on top of it.
type ManifestFormat string

const (
	ManifestFormatDocker ManifestFormat = "docker"
	ManifestFormatOCI    ManifestFormat = "oci"
)

func (format ManifestFormat) manifestMediaType() types.MediaType {
	if format == ManifestFormatOCI {
		return types.OCIManifestSchema1
	}
	return types.DockerManifestSchema2
}

func (format ManifestFormat) configMediaType() types.MediaType {
	if format == ManifestFormatOCI {
		return types.OCIConfigJSON
	}
	return types.DockerConfigJSON
}

func (format ManifestFormat) emptyLayer() *uncompressedLayer {
	if format == ManifestFormatOCI {
		return ociEmptyUncompressedLayer
	}
	return dockerEmptyUncompressedLayer
}

type manifestOnlyImage struct {
	CreatedAt time.Time
	Labels    map[string]string
	Format    ManifestFormat
}

func NewManifestOnlyImage(labels map[string]string, format ManifestFormat) v1.Image {
	img, err := newManifestOnlyImage(labels, format)
	if err != nil {
		panic(fmt.Sprintf("unable to create new ManifestOnlyImage: %s", err))
	}

	return img
}

func newManifestOnlyImage(labels map[string]string, format ManifestFormat) (v1.Image, error) {
	t := time.Now()

	img, err := partial.UncompressedToImage(manifestOnlyImage{
		CreatedAt: t,
		Labels:    labels,
		Format:    format,
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

	// partial.UncompressedToImage hardcodes the Docker schema2 manifest and config media types
	// regardless of MediaType() below, so they have to be set explicitly.
	img = mutate.MediaType(img, format.manifestMediaType())
	img = mutate.ConfigMediaType(img, format.configMediaType())

	return img, nil
}

// MediaType implements partial.UncompressedImageCore.
func (i manifestOnlyImage) MediaType() (types.MediaType, error) {
	return i.Format.manifestMediaType(), nil
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
			DiffIDs: []v1.Hash{i.Format.emptyLayer().diffID},
		},
	}, nil
}

func (i manifestOnlyImage) LayerByDiffID(h v1.Hash) (partial.UncompressedLayer, error) {
	return i.Format.emptyLayer(), nil
}
