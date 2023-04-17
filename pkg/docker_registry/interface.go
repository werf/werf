package docker_registry

import (
	"context"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/werf/pkg/image"
)

type Interface interface {
	CreateRepo(ctx context.Context, reference string) error
	DeleteRepo(ctx context.Context, reference string) error
	Tags(ctx context.Context, reference string, opts ...Option) ([]string, error)
	IsTagExist(ctx context.Context, reference string, opts ...Option) (bool, error)
	TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) error
	GetRepoImage(ctx context.Context, reference string) (*image.Info, error)
	TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error)
	DeleteRepoImage(ctx context.Context, repoImage *image.Info) error
	PushImage(ctx context.Context, reference string, opts *PushImageOptions) error
	MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateConfigFunc func(v1.Config) (v1.Config, error)) error
	CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) error

	PushImageArchive(ctx context.Context, archiveOpener ArchiveOpener, reference string) error
	PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error
	PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) error

	String() string

	parseReferenceParts(reference string) (referenceParts, error)
}

type ApiInterface interface {
	GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error)
}

type ArchiveOpener interface {
	Open() (io.ReadCloser, error)
}

type ManifestListOptions struct {
	PushImageOptions
	Manifests []*image.Info
}

type CopyImageOptions struct {
	IsImageIndex bool
}
