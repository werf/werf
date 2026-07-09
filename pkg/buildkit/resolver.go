package buildkit

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	"github.com/opencontainers/go-digest"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

var _ sourceresolver.ImageMetaResolver = (*ImageMetaResolver)(nil)

type ImageMetaResolver struct {
	defaultPlatform *ocispecs.Platform

	mu    sync.Mutex
	cache map[string]resolvedImage
}

type resolvedImage struct {
	ref    string
	digest digest.Digest
	config []byte
}

func NewImageMetaResolver(defaultPlatform *ocispecs.Platform) *ImageMetaResolver {
	return &ImageMetaResolver{
		defaultPlatform: defaultPlatform,
		cache:           map[string]resolvedImage{},
	}
}

func (r *ImageMetaResolver) ResolveImageConfig(ctx context.Context, ref string, opt sourceresolver.Opt) (string, digest.Digest, []byte, error) {
	platform := r.defaultPlatform
	if opt.ImageOpt != nil && opt.ImageOpt.Platform != nil {
		platform = opt.ImageOpt.Platform
	}

	cacheKey := ref
	if platform != nil {
		cacheKey = fmt.Sprintf("%s|%s/%s/%s", ref, platform.OS, platform.Architecture, platform.Variant)
	}

	r.mu.Lock()
	cached, ok := r.cache[cacheKey]
	r.mu.Unlock()
	if ok {
		return cached.ref, cached.digest, cached.config, nil
	}

	parsedRef, err := name.ParseReference(ref)
	if err != nil {
		return "", "", nil, fmt.Errorf("parse image reference %q: %w", ref, err)
	}

	remoteOpts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}
	if platform != nil {
		remoteOpts = append(remoteOpts, remote.WithPlatform(v1.Platform{
			OS:           platform.OS,
			Architecture: platform.Architecture,
			Variant:      platform.Variant,
		}))
	}

	desc, err := remote.Get(parsedRef, remoteOpts...)
	if err != nil {
		return "", "", nil, fmt.Errorf("get image %q from registry: %w", ref, err)
	}

	img, err := desc.Image()
	if err != nil {
		return "", "", nil, fmt.Errorf("get image %q manifest: %w", ref, err)
	}

	config, err := img.RawConfigFile()
	if err != nil {
		return "", "", nil, fmt.Errorf("get image %q config: %w", ref, err)
	}

	resolved := resolvedImage{
		ref:    ref,
		digest: digest.Digest(desc.Digest.String()),
		config: config,
	}

	r.mu.Lock()
	r.cache[cacheKey] = resolved
	r.mu.Unlock()

	return resolved.ref, resolved.digest, resolved.config, nil
}
