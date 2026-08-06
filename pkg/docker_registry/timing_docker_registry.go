package docker_registry

import (
	"context"
	"io"

	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ Interface = (*timingDockerRegistry)(nil)

// timingDockerRegistry records wall-clock durations of registry client calls
// into the opstats collector bound to the context, one operation per method.
type timingDockerRegistry struct {
	Interface
}

func newTimingDockerRegistry(registry Interface) *timingDockerRegistry {
	return &timingDockerRegistry{Interface: registry}
}

func observeRegistry(ctx context.Context, method string) func() {
	return opstats.Observe(ctx, opstats.Operation("registry: "+method))
}

func (r *timingDockerRegistry) CreateRepo(ctx context.Context, reference string) error {
	defer observeRegistry(ctx, "CreateRepo")()
	return r.Interface.CreateRepo(ctx, reference)
}

func (r *timingDockerRegistry) DeleteRepo(ctx context.Context, reference string) error {
	defer observeRegistry(ctx, "DeleteRepo")()
	return r.Interface.DeleteRepo(ctx, reference)
}

func (r *timingDockerRegistry) Tags(ctx context.Context, reference string, opts ...Option) ([]string, error) {
	defer observeRegistry(ctx, "Tags")()
	return r.Interface.Tags(ctx, reference, opts...)
}

func (r *timingDockerRegistry) IsTagExist(ctx context.Context, reference string, opts ...Option) (bool, error) {
	defer observeRegistry(ctx, "IsTagExist")()
	return r.Interface.IsTagExist(ctx, reference, opts...)
}

func (r *timingDockerRegistry) TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) error {
	defer observeRegistry(ctx, "TagRepoImage")()
	return r.Interface.TagRepoImage(ctx, repoImage, tag)
}

func (r *timingDockerRegistry) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	defer observeRegistry(ctx, "GetRepoImage")()
	return r.Interface.GetRepoImage(ctx, reference)
}

func (r *timingDockerRegistry) TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	defer observeRegistry(ctx, "TryGetRepoImage")()
	return r.Interface.TryGetRepoImage(ctx, reference)
}

func (r *timingDockerRegistry) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	defer observeRegistry(ctx, "DeleteRepoImage")()
	return r.Interface.DeleteRepoImage(ctx, repoImage)
}

func (r *timingDockerRegistry) PushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	defer observeRegistry(ctx, "PushImage")()
	return r.Interface.PushImage(ctx, reference, opts)
}

func (r *timingDockerRegistry) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, opts ...registry_api.MutateOption) error {
	defer observeRegistry(ctx, "MutateAndPushImage")()
	return r.Interface.MutateAndPushImage(ctx, sourceReference, destinationReference, opts...)
}

func (r *timingDockerRegistry) CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) error {
	defer observeRegistry(ctx, "CopyImage")()
	return r.Interface.CopyImage(ctx, sourceReference, destinationReference, opts)
}

func (r *timingDockerRegistry) PushImageArchive(ctx context.Context, archiveOpener ArchiveOpener, reference string) error {
	defer observeRegistry(ctx, "PushImageArchive")()
	return r.Interface.PushImageArchive(ctx, archiveOpener, reference)
}

func (r *timingDockerRegistry) PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error {
	defer observeRegistry(ctx, "PullImageArchive")()
	return r.Interface.PullImageArchive(ctx, archiveWriter, reference)
}

func (r *timingDockerRegistry) PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) error {
	defer observeRegistry(ctx, "PushManifestList")()
	return r.Interface.PushManifestList(ctx, reference, opts)
}
