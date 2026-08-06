package container_registry_extensions

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

var _ partial.UncompressedLayer = (*uncompressedLayer)(nil)

var EmptyUncompressedLayer = newEmptyUncompressedLayer()

// newEmptyUncompressedLayer returns a layer whose content is a valid empty tar archive.
// A zero-byte body is not a parseable tar: registries accept it, but consumers such as
// dive or docker save readers fail with EOF on every image that inherits the layer.
func newEmptyUncompressedLayer() *uncompressedLayer {
	var buf bytes.Buffer
	if err := tar.NewWriter(&buf).Close(); err != nil {
		panic(fmt.Sprintf("write empty tar archive: %s", err))
	}

	diffID, _, err := v1.SHA256(bytes.NewReader(buf.Bytes()))
	if err != nil {
		panic(fmt.Sprintf("calculate empty tar archive diffID: %s", err))
	}

	return &uncompressedLayer{
		diffID:    diffID,
		mediaType: types.DockerLayer,
		content:   buf.Bytes(),
	}
}

type uncompressedLayer struct {
	diffID    v1.Hash
	mediaType types.MediaType
	content   []byte
}

func (layer *uncompressedLayer) DiffID() (v1.Hash, error) {
	return layer.diffID, nil
}

func (layer *uncompressedLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(layer.content)), nil
}

func (layer *uncompressedLayer) MediaType() (types.MediaType, error) {
	return layer.mediaType, nil
}
