package docker_registry

import (
	"context"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

type DockerRegistryTracer struct {
	DockerRegistry    Interface
	DockerRegistryApi GenericApiInterface
}

func NewDockerRegistryTracer(dockerRegistry Interface, dockerRegistryApi GenericApiInterface) *DockerRegistryTracer {
	return &DockerRegistryTracer{
		DockerRegistry:    dockerRegistry,
		DockerRegistryApi: dockerRegistryApi,
	}
}

func (r *DockerRegistryTracer) CreateRepo(ctx context.Context, reference string) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.CreateRepo %q", reference).Do(func() {
		err = r.DockerRegistry.CreateRepo(ctx, reference)
	})
	return
}

func (r *DockerRegistryTracer) DeleteRepo(ctx context.Context, reference string) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.DeleteRepo %q", reference).Do(func() {
		err = r.DockerRegistry.DeleteRepo(ctx, reference)
	})
	return
}

func (r *DockerRegistryTracer) Tags(ctx context.Context, reference string, opts ...Option) (res []string, err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.Tags %q", reference).Do(func() {
		res, err = r.DockerRegistry.Tags(ctx, reference, opts...)
	})
	return
}

func (r *DockerRegistryTracer) IsTagExist(ctx context.Context, reference string, opts ...Option) (res bool, err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.IsTagExist %q", reference).Do(func() {
		res, err = r.DockerRegistry.IsTagExist(ctx, reference, opts...)
	})
	return
}

func (r *DockerRegistryTracer) TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.TagRepoImage %q", tag).Do(func() {
		err = r.DockerRegistry.TagRepoImage(ctx, repoImage, tag)
	})
	return
}

func (r *DockerRegistryTracer) GetRepoImage(ctx context.Context, reference string) (res *image.Info, err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.GetRepoImage %q", reference).Do(func() {
		if r.DockerRegistry != nil {
			res, err = r.DockerRegistry.GetRepoImage(ctx, reference)
		} else {
			res, err = r.DockerRegistryApi.GetRepoImage(ctx, reference)
		}
	})
	return
}

func (r *DockerRegistryTracer) TryGetRepoImage(ctx context.Context, reference string) (res *image.Info, err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.TryGetRepoImage %q", reference).Do(func() {
		res, err = r.DockerRegistry.TryGetRepoImage(ctx, reference)
	})
	return
}

func (r *DockerRegistryTracer) DeleteRepoImage(ctx context.Context, repoImage *image.Info) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.DeleteRepoImage %v", repoImage).Do(func() {
		err = r.DockerRegistry.DeleteRepoImage(ctx, repoImage)
	})
	return
}

func (r *DockerRegistryTracer) PushImage(ctx context.Context, reference string, opts *PushImageOptions) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.PushImage %q", reference).Do(func() {
		err = r.DockerRegistry.PushImage(ctx, reference, opts)
	})
	return
}

func (r *DockerRegistryTracer) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateConfigFunc func(v1.Config) (v1.Config, error)) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.MutateAndPushImage %q -> %q", sourceReference, destinationReference).Do(func() {
		if r.DockerRegistry != nil {
			err = r.DockerRegistry.MutateAndPushImage(ctx, sourceReference, destinationReference, mutateConfigFunc)
		} else {
			err = r.DockerRegistryApi.MutateAndPushImage(ctx, sourceReference, destinationReference, mutateConfigFunc)
		}
	})
	return
}

func (r *DockerRegistryTracer) CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.CopyImage %q -> %q", sourceReference, destinationReference).Do(func() {
		err = r.DockerRegistry.CopyImage(ctx, sourceReference, destinationReference, opts)
	})
	return
}

func (r *DockerRegistryTracer) PushImageArchive(ctx context.Context, archiveOpener ArchiveOpener, reference string) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.PushImageArchive %q", reference).Do(func() {
		err = r.DockerRegistry.PushImageArchive(ctx, archiveOpener, reference)
	})
	return
}

func (r *DockerRegistryTracer) PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.PullImageArchive %q", reference).Do(func() {
		err = r.DockerRegistry.PullImageArchive(ctx, archiveWriter, reference)
	})
	return
}

func (r *DockerRegistryTracer) PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) (err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.PushManifestList %q", reference).Do(func() {
		err = r.DockerRegistry.PushManifestList(ctx, reference, opts)
	})
	return
}

func (r *DockerRegistryTracer) String() (res string) {
	return r.DockerRegistry.String()
}

func (r *DockerRegistryTracer) parseReferenceParts(reference string) (res referenceParts, err error) {
	return r.DockerRegistry.parseReferenceParts(reference)
}

func (r *DockerRegistryTracer) GetRepoImageConfigFile(ctx context.Context, reference string) (res *v1.ConfigFile, err error) {
	logboek.Context(ctx).Default().LogProcess("DockerRegistryTracer.GetRepoImageConfigFile %q", reference).Do(func() {
		res, err = r.DockerRegistryApi.GetRepoImageConfigFile(ctx, reference)
	})
	return
}
