package api

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type MutateOption func(*mutateOptions)

type mutateOptions struct {
	mutateConfigFunc              func(context.Context, v1.Config) (v1.Config, error)
	mutateConfigFileFunc          func(context.Context, *v1.ConfigFile) (*v1.ConfigFile, error)
	mutateImageLayersFunc         func(context.Context, []v1.Layer) ([]mutate.Addendum, error)
	mutateManifestAnnotationsFunc func(context.Context, *v1.Manifest) (map[string]string, error)
}

func WithConfigMutation(f func(context.Context, v1.Config) (v1.Config, error)) MutateOption {
	return func(opts *mutateOptions) {
		opts.mutateConfigFunc = f
	}
}

func WithConfigFileMutation(f func(context.Context, *v1.ConfigFile) (*v1.ConfigFile, error)) MutateOption {
	return func(opts *mutateOptions) {
		opts.mutateConfigFileFunc = f
	}
}

func WithLayersMutation(f func(context.Context, []v1.Layer) ([]mutate.Addendum, error)) MutateOption {
	return func(opts *mutateOptions) {
		opts.mutateImageLayersFunc = f
	}
}

func WithManifestAnnotationsFunc(f func(context.Context, *v1.Manifest) (map[string]string, error)) MutateOption {
	return func(opts *mutateOptions) {
		opts.mutateManifestAnnotationsFunc = f
	}
}

type Api interface {
	MutateImageOrIndex(ctx context.Context, imageOrIndex interface{}, dest name.Reference, isDestRefByDigest bool, opts ...MutateOption) (interface{}, error)
}

type MutateImageOrIndexOpts struct {
	ImageOrIndex      interface{}
	Dest              name.Reference
	IsDestRefByDigest bool
	MutateOptions     []MutateOption
}

func MutateImageOrIndex(ctx context.Context, opts MutateImageOrIndexOpts) (interface{}, name.Reference, error) {
	switch obj := opts.ImageOrIndex.(type) {
	case v1.Image:
		return mutateImage(ctx, obj, opts.Dest, opts.IsDestRefByDigest, opts.MutateOptions...)
	case v1.ImageIndex:
		return mutateIndex(ctx, obj, opts.Dest, opts.IsDestRefByDigest, opts.MutateOptions...)
	default:
		return nil, nil, fmt.Errorf("unsupported type %T", opts.ImageOrIndex)
	}
}

func mutateImage(ctx context.Context, image v1.Image, dest name.Reference, isDestRefByDigest bool, opts ...MutateOption) (v1.Image, name.Reference, error) {
	options := applyMutateOptions(opts...)

	cf, err := image.ConfigFile()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading config file: %w", err)
	}

	manifest, err := image.Manifest()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading image manifest: %w", err)
	}

	if options.mutateConfigFileFunc != nil || options.mutateConfigFunc != nil {
		newCF := cf
		if options.mutateConfigFileFunc != nil {
			newCF, err = options.mutateConfigFileFunc(ctx, cf)
			if err != nil {
				return nil, nil, err
			}
		}
		if options.mutateConfigFunc != nil {
			newConfig, err := options.mutateConfigFunc(ctx, cf.Config)
			if err != nil {
				return nil, nil, err
			}
			newCF.Config = newConfig
		}
		image, err = mutate.ConfigFile(image, newCF)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to mutate image config: %w", err)
		}
	}

	if options.mutateImageLayersFunc != nil {
		layers, err := image.Layers()
		if err != nil {
			return nil, nil, err
		}
		newLayers, err := options.mutateImageLayersFunc(ctx, layers)
		if err != nil {
			return nil, nil, err
		}
		image, err = mutate.Append(empty.Image, newLayers...)
		if err != nil {
			return nil, nil, err
		}
		image, err = mutate.ConfigFile(image, cf)
		if err != nil {
			return nil, nil, err
		}

		// preserve manifest annotations
		image = mutate.Annotations(image, manifest.Annotations).(v1.Image)
	}

	if options.mutateManifestAnnotationsFunc != nil {
		manifestAnnotations, err := options.mutateManifestAnnotationsFunc(ctx, manifest)
		if err != nil {
			return nil, nil, fmt.Errorf("error mutating manifest annotations: %w", err)
		}

		image = mutate.Annotations(image, manifestAnnotations).(v1.Image)
	}

	digest, err := image.Digest()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get new image digest: %w", err)
	}

	destRef := resolveDestRef(dest, digest, isDestRefByDigest)

	return image, destRef, nil
}

func mutateIndex(ctx context.Context, index v1.ImageIndex, dest name.Reference, isDestRefByDigest bool, opts ...MutateOption) (v1.ImageIndex, name.Reference, error) {
	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, nil, fmt.Errorf("getting image index manifest: %w", err)
	}

	newIndex := mutate.IndexMediaType(empty.Index, types.DockerManifestList)
	addenda := make([]mutate.IndexAddendum, 0, len(indexManifest.Manifests))

	for _, desc := range indexManifest.Manifests {
		subRef := dest.Context().Digest(desc.Digest.String())
		addendum, err := mutateManifestEntry(ctx, index, desc, subRef, opts...)
		if err != nil {
			return nil, nil, err
		}
		addenda = append(addenda, *addendum)
	}

	newIndex = mutate.AppendManifests(newIndex, addenda...)

	digest, err := newIndex.Digest()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting new image index digest: %w", err)
	}

	destRef := resolveDestRef(dest, digest, isDestRefByDigest)

	return newIndex, destRef, nil
}

func mutateManifestEntry(ctx context.Context, index v1.ImageIndex, desc v1.Descriptor, dest name.Reference, opts ...MutateOption) (*mutate.IndexAddendum, error) {
	switch {
	case desc.MediaType.IsIndex():
		subIdx, err := index.ImageIndex(desc.Digest)
		if err != nil {
			return nil, fmt.Errorf("getting index by digest %s: %w", desc.Digest, err)
		}
		newIdx, _, err := mutateIndex(ctx, subIdx, dest, true, opts...)
		if err != nil {
			return nil, fmt.Errorf("mutating sub-index %s: %w", dest, err)
		}
		desc, err := partial.Descriptor(newIdx)
		if err != nil {
			return nil, fmt.Errorf("creating descriptor for index %s: %w", dest, err)
		}
		return &mutate.IndexAddendum{Add: newIdx, Descriptor: *desc}, nil

	case desc.MediaType.IsImage():
		subImg, err := index.Image(desc.Digest)
		if err != nil {
			return nil, fmt.Errorf("getting image by digest %s: %w", desc.Digest, err)
		}
		newImg, _, err := mutateImage(ctx, subImg, dest, true, opts...)
		if err != nil {
			return nil, fmt.Errorf("mutating sub-image %s: %w", dest, err)
		}
		cf, err := newImg.ConfigFile()
		if err != nil {
			return nil, fmt.Errorf("getting config file of %s: %w", dest, err)
		}
		desc, err := partial.Descriptor(newImg)
		if err != nil {
			return nil, fmt.Errorf("creating descriptor for image %s: %w", dest, err)
		}
		desc.Platform = cf.Platform()
		return &mutate.IndexAddendum{Add: newImg, Descriptor: *desc}, nil

	default:
		return nil, fmt.Errorf("unsupported media type %q", desc.MediaType)
	}
}

func resolveDestRef(dest name.Reference, digest v1.Hash, byDigest bool) name.Reference {
	if byDigest {
		return dest.Context().Digest(digest.String())
	}
	return dest
}

func applyMutateOptions(opts ...MutateOption) *mutateOptions {
	options := &mutateOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}
