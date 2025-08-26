package container_backend

import (
	"context"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/werf/werf/v2/pkg/image"
)

type BackendLoaderStorer interface {
	SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error)
	LoadImageFromStream(ctx context.Context, input io.Reader) (string, error)
}

func MutateAndPushImage(ctx context.Context, imageName string, newConfig image.SpecConfig, backend BackendLoaderStorer) (string, error) {
	opener := func() (io.ReadCloser, error) {
		return backend.SaveImageToStream(ctx, imageName)
	}

	img, err := tarball.Image(opener, nil)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	config, err := img.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("error set config: %w", err)
	}

	image.UpdateConfigFile(newConfig, config)

	img, err = mutate.ConfigFile(img, config)
	if err != nil {
		return "", fmt.Errorf("error mutate config: %w", err)
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := tarball.Write(nil, img, pw); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to write tarball: %w", err))
		}
	}()

	newDigest, err := backend.LoadImageFromStream(ctx, pr)
	if err != nil {
		return "", fmt.Errorf("unable to load image: %w", err)
	}

	return newDigest, nil
}
