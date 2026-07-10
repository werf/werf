package container_backend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

// NativeConfigMutator is implemented by backends that can mutate an image's config via a native
// build+commit flow (create container from src, apply config, commit to dest), avoiding the
// save-to-tar/load-from-tar roundtrip that MutateAndPushImage otherwise requires.
type NativeConfigMutator interface {
	MutateAndPushImageNative(ctx context.Context, src, dest string, newConfig image.SpecConfig, targetPlatform string) error
}

// ErrNativeMutationUnsupported signals that a backend cannot faithfully perform the requested
// mutation via its native path and the caller must fall back to MutateAndPushImage (save/load).
var ErrNativeMutationUnsupported = errors.New("native mutation not supported for this config")

// IsNativeMutationUnsupported reports whether err indicates the native mutation path declined the
// mutation and the save/load fallback must be used.
func IsNativeMutationUnsupported(err error) bool {
	return errors.Is(err, ErrNativeMutationUnsupported)
}

// nativeMutationEligible reports whether newConfig can be faithfully reproduced by docker commit,
// whose daemon-side merge is additive-only: it re-adds base labels/env/volumes for omitted keys,
// cannot clear fields, and cannot drop history. The mutation therefore matches werf's full-replace
// UpdateConfigFile semantics only when it neither clears anything nor removes any base
// label/env/volume/exposed-port.
func nativeMutationEligible(newConfig image.SpecConfig, baseConfig *v1.ConfigFile) bool {
	if newConfig.ClearHistory || newConfig.ClearCmd || newConfig.ClearEntrypoint || newConfig.ClearUser || newConfig.ClearWorkingDir {
		return false
	}

	base := baseConfig.Config

	// docker create refuses images with no command ("no command specified"), so the native
	// create+commit path can't handle a config whose effective Cmd and Entrypoint are both empty.
	// A slice of only empty strings (e.g. Entrypoint [""]) counts as no command to the daemon.
	if isEmptyCommand(newConfig.Cmd) && isEmptyCommand(base.Cmd) && isEmptyCommand(newConfig.Entrypoint) && isEmptyCommand(base.Entrypoint) {
		return false
	}

	if newConfig.Labels != nil && !isMapSuperset(newConfig.Labels, base.Labels) {
		return false
	}
	if newConfig.Volumes != nil && !isSetSuperset(newConfig.Volumes, base.Volumes) {
		return false
	}
	if newConfig.ExposedPorts != nil && !isSetSuperset(newConfig.ExposedPorts, base.ExposedPorts) {
		return false
	}
	// Env is replaced unconditionally by UpdateConfigFile (even when nil clears it), so the
	// superset check must run regardless of whether newConfig.Env is nil.
	if !isEnvSuperset(newConfig.Env, base.Env) {
		return false
	}

	return true
}

func isEmptyCommand(cmd []string) bool {
	for _, c := range cmd {
		if c != "" {
			return false
		}
	}
	return true
}

func isMapSuperset(next, base map[string]string) bool {
	for k := range base {
		if _, ok := next[k]; !ok {
			return false
		}
	}
	return true
}

func isSetSuperset(next, base map[string]struct{}) bool {
	for k := range base {
		if _, ok := next[k]; !ok {
			return false
		}
	}
	return true
}

func isEnvSuperset(next, base []string) bool {
	nextKeys := make(map[string]struct{}, len(next))
	for _, entry := range next {
		key, _, _ := strings.Cut(entry, "=")
		nextKeys[key] = struct{}{}
	}
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if _, ok := nextKeys[key]; !ok {
			return false
		}
	}
	return true
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
