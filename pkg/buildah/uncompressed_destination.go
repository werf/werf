package buildah

import (
	"context"

	"go.podman.io/image/v5/types"
)

// uncompressedDestinationReference makes a registry destination publish uncompressed layers:
// containers/image registry destinations always ask for compressed layers and offer no option to
// turn it off.
type uncompressedDestinationReference struct {
	types.ImageReference
}

func lookupUncompressedDestination(ref types.ImageReference) (types.ImageReference, error) {
	return uncompressedDestinationReference{ImageReference: ref}, nil
}

func (ref uncompressedDestinationReference) NewImageDestination(ctx context.Context, sys *types.SystemContext) (types.ImageDestination, error) {
	dest, err := ref.ImageReference.NewImageDestination(ctx, sys)
	if err != nil {
		return nil, err
	}

	return uncompressedDestination{ImageDestination: dest}, nil
}

type uncompressedDestination struct {
	types.ImageDestination
}

func (dest uncompressedDestination) DesiredLayerCompression() types.LayerCompression {
	return types.Decompress
}

// TryReusingBlob only reuses the exact uncompressed blob: a substitute is a compressed variant of the
// layer known to the blob info cache, which DesiredLayerCompression does not guard against.
func (dest uncompressedDestination) TryReusingBlob(ctx context.Context, info types.BlobInfo, cache types.BlobInfoCache, _ bool) (bool, types.BlobInfo, error) {
	return dest.ImageDestination.TryReusingBlob(ctx, info, cache, false)
}
