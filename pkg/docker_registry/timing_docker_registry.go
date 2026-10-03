package docker_registry

import (
	"context"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	registry_api "github.com/werf/werf/v3/pkg/docker_registry/api"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
)

var _ Interface = (*timingDockerRegistry)(nil)

// timingDockerRegistry records wall-clock durations of registry client calls
// into the opstats collector bound to the context, one operation per method.
// The names measure client calls, not individual HTTP requests.
type timingDockerRegistry struct {
	Interface
}

func newTimingDockerRegistry(registry Interface) *timingDockerRegistry {
	return &timingDockerRegistry{Interface: registry}
}

func observeRegistry(ctx context.Context, operation string) func() {
	return opstats.Observe(ctx, opstats.Operation("registry: "+operation))
}

func (r *timingDockerRegistry) CreateRepo(ctx context.Context, reference string) error {
	defer observeRegistry(ctx, "repository create")()
	return r.Interface.CreateRepo(ctx, reference)
}

func (r *timingDockerRegistry) DeleteRepo(ctx context.Context, reference string) error {
	defer observeRegistry(ctx, "repository delete")()
	return r.Interface.DeleteRepo(ctx, reference)
}

func (r *timingDockerRegistry) Tags(ctx context.Context, reference string, opts ...Option) ([]string, error) {
	defer observeRegistry(ctx, "tags list")()
	return r.Interface.Tags(ctx, reference, opts...)
}

func (r *timingDockerRegistry) IsTagExist(ctx context.Context, reference string, opts ...Option) (bool, error) {
	defer observeRegistry(ctx, "tag exists")()
	return r.Interface.IsTagExist(ctx, reference, opts...)
}

func (r *timingDockerRegistry) TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) error {
	defer observeRegistry(ctx, "image tag")()
	return r.Interface.TagRepoImage(ctx, repoImage, tag)
}

func (r *timingDockerRegistry) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	defer observeRegistry(ctx, "image get")()
	return r.Interface.GetRepoImage(ctx, reference)
}

func (r *timingDockerRegistry) TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	defer observeRegistry(ctx, "image try get")()
	return r.Interface.TryGetRepoImage(ctx, reference)
}

func (r *timingDockerRegistry) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	defer observeRegistry(ctx, "image delete")()
	return r.Interface.DeleteRepoImage(ctx, repoImage)
}

func (r *timingDockerRegistry) PushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	defer observeRegistry(ctx, "image push")()
	return r.Interface.PushImage(ctx, reference, opts)
}

func (r *timingDockerRegistry) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, opts ...registry_api.MutateOption) error {
	defer observeRegistry(ctx, "image mutate and push")()
	return r.Interface.MutateAndPushImage(ctx, sourceReference, destinationReference, opts...)
}

func (r *timingDockerRegistry) CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) error {
	defer observeRegistry(ctx, "image copy")()
	return r.Interface.CopyImage(ctx, sourceReference, destinationReference, opts)
}

func (r *timingDockerRegistry) PushImageArchive(ctx context.Context, archiveOpener ArchiveOpener, reference string) error {
	defer observeRegistry(ctx, "image archive push")()
	return r.Interface.PushImageArchive(ctx, archiveOpener, reference)
}

func (r *timingDockerRegistry) PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error {
	defer observeRegistry(ctx, "image archive pull")()
	return r.Interface.PullImageArchive(ctx, archiveWriter, reference)
}

func (r *timingDockerRegistry) PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) error {
	defer observeRegistry(ctx, "manifest list push")()
	return r.Interface.PushManifestList(ctx, reference, opts)
}

var _ GenericApiInterface = (*timingGenericApi)(nil)

// timingGenericApi is the timing decorator for the shared generic registry API
// (base-image lookups, dependency fetches) returned by API().
type timingGenericApi struct {
	GenericApiInterface
}

func newTimingGenericApi(api GenericApiInterface) *timingGenericApi {
	return &timingGenericApi{GenericApiInterface: api}
}

func (r *timingGenericApi) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	defer observeRegistry(ctx, "image get")()
	return r.GenericApiInterface.GetRepoImage(ctx, reference)
}

func (r *timingGenericApi) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, opts ...registry_api.MutateOption) error {
	defer observeRegistry(ctx, "image mutate and push")()
	return r.GenericApiInterface.MutateAndPushImage(ctx, sourceReference, destinationReference, opts...)
}

func (r *timingGenericApi) GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error) {
	defer observeRegistry(ctx, "image config get")()
	return r.GenericApiInterface.GetRepoImageConfigFile(ctx, reference)
}
