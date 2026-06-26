package container_backend

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd/platforms"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

type BackendLoaderStorer interface {
	SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error)
	LoadImageFromStream(ctx context.Context, input io.Reader) (string, error)
}

type ImageConfigFileGetter interface {
	GetImageConfigFile(ctx context.Context, imageName string) (*v1.ConfigFile, error)
}

func MutateAndPushImage(ctx context.Context, imageName, targetPlatform string, newConfig image.SpecConfig, backend BackendLoaderStorer) (string, error) {
	if getter, ok := backend.(ImageConfigFileGetter); ok {
		config, err := getter.GetImageConfigFile(ctx, imageName)
		if err != nil {
			return "", fmt.Errorf("failed to get image config: %w", err)
		}
		if len(config.RootFS.DiffIDs) == 0 {
			return mutateAndLoadLayerlessImage(ctx, targetPlatform, newConfig, config, backend)
		}
	}

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

	if err := applyConfigMutations(targetPlatform, newConfig, config); err != nil {
		return "", err
	}

	img, err = mutate.ConfigFile(img, config)
	if err != nil {
		return "", fmt.Errorf("error mutate config: %w", err)
	}

	return writeAndLoadImage(ctx, img, backend)
}

func mutateAndLoadLayerlessImage(ctx context.Context, targetPlatform string, newConfig image.SpecConfig, config *v1.ConfigFile, backend BackendLoaderStorer) (string, error) {
	if err := applyConfigMutations(targetPlatform, newConfig, config); err != nil {
		return "", err
	}

	img, err := mutate.ConfigFile(empty.Image, config)
	if err != nil {
		return "", fmt.Errorf("error mutate config: %w", err)
	}

	return writeAndLoadImage(ctx, img, backend)
}

func applyConfigMutations(targetPlatform string, newConfig image.SpecConfig, config *v1.ConfigFile) error {
	if targetPlatform != "" {
		platformSpec, err := platforms.Parse(targetPlatform)
		if err != nil {
			return fmt.Errorf("parse target platform %q: %w", targetPlatform, err)
		}

		config.OS = platformSpec.OS
		config.Architecture = platformSpec.Architecture
		config.Variant = platformSpec.Variant
	}

	image.UpdateConfigFile(newConfig, config)
	return nil
}

func writeAndLoadImage(ctx context.Context, img v1.Image, backend BackendLoaderStorer) (string, error) {
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
