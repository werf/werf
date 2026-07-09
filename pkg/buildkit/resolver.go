package buildkit

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
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
	ImageMetaResolverOptions

	mu    sync.Mutex
	cache map[string]resolvedImage
}

type ImageMetaResolverOptions struct {
	InsecureRegistry      bool
	SkipTLSVerifyRegistry bool
}

type resolvedImage struct {
	ref    string
	digest digest.Digest
	config []byte
}

func NewImageMetaResolver(defaultPlatform *ocispecs.Platform, opts ImageMetaResolverOptions) *ImageMetaResolver {
	return &ImageMetaResolver{
		defaultPlatform:          defaultPlatform,
		ImageMetaResolverOptions: opts,
		cache:                    map[string]resolvedImage{},
	}
}

func (r *ImageMetaResolver) parseReferenceOptions() []name.Option {
	if r.InsecureRegistry {
		return []name.Option{name.Insecure}
	}
	return nil
}

// ResolvePinnedRef resolves ref and returns it pinned to the resolved digest
// ("<repo>@sha256:…") together with the image config.
func (r *ImageMetaResolver) ResolvePinnedRef(ctx context.Context, ref string, platform *ocispecs.Platform) (string, []byte, error) {
	_, dgst, config, err := r.ResolveImageConfig(ctx, ref, sourceresolver.Opt{
		ImageOpt: &sourceresolver.ResolveImageOpt{Platform: platform},
	})
	if err != nil {
		return "", nil, err
	}

	parsedRef, err := name.ParseReference(ref, r.parseReferenceOptions()...)
	if err != nil {
		return "", nil, fmt.Errorf("parse image reference %q: %w", ref, err)
	}

	return fmt.Sprintf("%s@%s", parsedRef.Context().Name(), dgst), config, nil
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

	parsedRef, err := name.ParseReference(ref, r.parseReferenceOptions()...)
	if err != nil {
		return "", "", nil, fmt.Errorf("parse image reference %q: %w", ref, err)
	}

	remoteOpts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}
	if r.SkipTLSVerifyRegistry {
		transport := remote.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		remoteOpts = append(remoteOpts, remote.WithTransport(transport))
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
