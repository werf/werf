package container_registry_extensions

import (
	"bytes"
	"io"
	"io/ioutil"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

var EmptyUncompressedLayer = &uncompressedLayer{
	diffID: v1.Hash{
		Algorithm: "sha256",
		Hex:       "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	},
	mediaType: types.DockerLayer,
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
	return ioutil.NopCloser(bytes.NewBuffer(layer.content)), nil
}

func (layer *uncompressedLayer) MediaType() (types.MediaType, error) {
	return layer.mediaType, nil
}
