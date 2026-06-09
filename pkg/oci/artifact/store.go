package artifact

import (
	"context"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
)

type Store interface {
	Attach(ctx context.Context, parentDigest, artifactType string, payload []byte, checksum, targetPlatform string) error
	GetAttachedContent(ctx context.Context, parentDigest, artifactType string) ([]byte, error)
	GetAttachedContentAny(ctx context.Context, parentDigest, artifactType string) ([]byte, error)
	GetAttached(ctx context.Context, parentDigest, artifactType string) (v1.Descriptor, bool, error)
}

// OCIStore manages OCI artifacts attached to container images.
// An artifact is an OCI image with a subject reference pointing to the parent image digest.
type OCIStore struct {
	repo      string
	imageName string
	opts      []remote.Option
}

// NewOCIStore creates a new OCIStore for the given repository and optional image name.
// By default, registry authentication is handled via the global docker_registry API.
// Explicit remote options can be passed to override the default auth, though in most
// cases the default werf registry authentication should suffice.
func NewOCIStore(repo, imageName string, opts ...remote.Option) *OCIStore {
	return &OCIStore{
		repo:      repo,
		imageName: imageName,
		opts:      opts,
	}
}

func (s *OCIStore) Attach(ctx context.Context, parentDigest, artifactType string, payload []byte, checksum, targetPlatform string) error {
	layer := static.NewLayer(payload, types.MediaType(artifactType))
	img, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return fmt.Errorf("append artifact layer: %w", err)
	}
	img = mutate.MediaType(img, types.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, types.MediaType(EmptyConfigMediaType))

	parentRef, err := name.NewDigest(s.repo + "@" + parentDigest)
	if err != nil {
		return fmt.Errorf("parse parent digest reference: %w", err)
	}
	parentDesc, err := remote.Get(parentRef, s.remoteOptions(ctx)...)
	if err != nil {
		return fmt.Errorf("get parent descriptor: %w", err)
	}
	imgWithSubject := mutate.Subject(img, parentDesc.Descriptor).(v1.Image)

	if err := PushArtifactImage(ctx, s.repo, imgWithSubject, s.remoteOptions(ctx)...); err != nil {
		return err
	}

	desc, err := partial.Descriptor(imgWithSubject)
	if err != nil {
		return fmt.Errorf("create descriptor: %w", err)
	}
	artifactDesc := *desc
	artifactDesc.ArtifactType = artifactType

	annotations := make(map[string]string)
	if checksum != "" {
		annotations[image.WerfChecksumAnnotation] = checksum
	}
	if s.imageName != "" {
		annotations[image.WerfImageNameAnnotation] = s.imageName
	}
	if targetPlatform != "" {
		annotations[image.WerfPlatformAnnotation] = targetPlatform
	}
	if len(annotations) > 0 {
		artifactDesc.Annotations = annotations
	}

	return Attach(ctx, s.repo, parentDigest, artifactDesc, artifactType, s.imageName, s.remoteOptions(ctx)...)
}

// GetAttached returns the descriptor of an artifact attached to the given parent image digest.
//
// Two parent images with the same digest share the same attached artifact: if image A
// at digest D has an SBOM attached, any image with digest D (same content) shares that SBOM.
// The artifact is identified by the (parentDigest, artifactType, imageName) tuple.
func (s *OCIStore) GetAttached(ctx context.Context, parentDigest, artifactType string) (v1.Descriptor, bool, error) {
	return GetAttached(ctx, s.repo, parentDigest, artifactType, s.imageName, s.remoteOptions(ctx)...)
}

func (s *OCIStore) GetAttachedContent(ctx context.Context, parentDigest, artifactType string) ([]byte, error) {
	desc, found, err := s.GetAttached(ctx, parentDigest, artifactType)
	if err != nil {
		return nil, fmt.Errorf("get attached artifact: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no artifact of type %q found for digest %q: %w", artifactType, parentDigest, ErrNotFound)
	}

	return s.pullLayerContent(ctx, desc.Digest.String())
}

// GetAttachedContentAny returns the content of the first matching artifact
// attached to the given parent digest, regardless of image name. If multiple
// artifacts of the same type exist (e.g., different images sharing the same
// parent digest), the first match is returned and a warning is logged.
// Callers needing a specific artifact should use GetAttachedContent with an
// imageName-configured store instead.
func (s *OCIStore) GetAttachedContentAny(ctx context.Context, parentDigest, artifactType string) ([]byte, error) {
	desc, found, err := GetAttached(ctx, s.repo, parentDigest, artifactType, "", s.opts...)
	if err != nil {
		return nil, fmt.Errorf("get attached artifact: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no artifact of type %q found for digest %q: %w", artifactType, parentDigest, ErrNotFound)
	}

	return s.pullLayerContent(ctx, desc.Digest.String())
}

func (s *OCIStore) pullLayerContent(ctx context.Context, digest string) ([]byte, error) {
	imageRef, err := name.NewDigest(s.repo + "@" + digest)
	if err != nil {
		return nil, fmt.Errorf("parse artifact digest reference: %w", err)
	}

	img, err := remote.Image(imageRef, s.remoteOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("pull artifact image: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("get artifact layers: %w", err)
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("artifact has no layers")
	}

	payload, err := readLayerContent(layers[0])
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func readLayerContent(layer v1.Layer) ([]byte, error) {
	rc, err := layer.Compressed()
	if err != nil {
		return nil, fmt.Errorf("read artifact layer: %w", err)
	}
	defer rc.Close()

	payload, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read artifact content: %w", err)
	}

	return payload, nil
}

func (s *OCIStore) remoteOptions(ctx context.Context) []remote.Option {
	opts := s.opts
	if len(opts) == 0 {
		opts = docker_registry.API().RemoteOptionsForHost(ctx, s.repo)
	}
	return append([]remote.Option{remote.WithContext(ctx)}, opts...)
}
