package container_backend

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

type BackendLoaderStorer interface {
	SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error)
	LoadImageFromStream(ctx context.Context, input io.Reader) (string, error)
}

func MutateAndPushImage(ctx context.Context, imageName string, newConfig image.SpecConfig, backend BackendLoaderStorer) (string, error) {
	imageStream, err := backend.SaveImageToStream(ctx, imageName)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	tempTarFile, err := tmp_manager.TempFile("werf-image-mutation-*.tar")
	if err != nil {
		_ = imageStream.Close()
		return "", fmt.Errorf("failed to create temp tar file: %w", err)
	}
	tempTarPath := tempTarFile.Name()
	defer func() {
		_ = os.Remove(tempTarPath)
	}()

	if _, err := io.Copy(tempTarFile, imageStream); err != nil {
		_ = imageStream.Close()
		_ = tempTarFile.Close()
		return "", fmt.Errorf("failed to persist image tarball: %w", err)
	}

	if err := imageStream.Close(); err != nil {
		_ = tempTarFile.Close()
		return "", fmt.Errorf("failed to close image save stream: %w", err)
	}

	if err := tempTarFile.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize temp tar file: %w", err)
	}

	img, err := tarball.ImageFromPath(tempTarPath, nil)
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
		defer func() {
			_ = pw.Close()
		}()
		if err := tarball.Write(nil, img, pw); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("failed to write tarball: %w", err))
		}
	}()

	newDigest, err := backend.LoadImageFromStream(ctx, pr)
	if err != nil {
		return "", fmt.Errorf("unable to load image: %w", err)
	}

	return newDigest, nil
}
