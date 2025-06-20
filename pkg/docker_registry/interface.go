package docker_registry

import (
	"context"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/werf/werf/v2/pkg/image"
)

type commonInterface interface {
	GetRepoImage(ctx context.Context, reference string) (*image.Info, error)
	MutateAndPushImageLayers(ctx context.Context, sourceReference, destinationReference string, f func(context.Context, []v1.Layer) ([]mutate.Addendum, error)) error
	MutateAndPushImageConfigFile(ctx context.Context, sourceReference, destinationReference string, f func(context.Context, *v1.ConfigFile) (*v1.ConfigFile, error)) error
	MutateAndPushImageConfig(ctx context.Context, sourceReference, destinationReference string, f func(context.Context, v1.Config) (v1.Config, error)) error
}

type Interface interface {
	commonInterface

	CreateRepo(ctx context.Context, reference string) error
	DeleteRepo(ctx context.Context, reference string) error
	Tags(ctx context.Context, reference string, opts ...Option) ([]string, error)
	IsTagExist(ctx context.Context, reference string, opts ...Option) (bool, error)
	TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) error
	TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error)
	DeleteRepoImage(ctx context.Context, repoImage *image.Info) error
	PushImage(ctx context.Context, reference string, opts *PushImageOptions) error
	CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) error

	PushImageArchive(ctx context.Context, archiveOpener ArchiveOpener, reference string) error
	PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error
	PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) error

	String() string

	parseReferenceParts(reference string) (referenceParts, error)
}

type GenericApiInterface interface {
	commonInterface

	GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error)
}

type ArchiveOpener interface {
	Open() (io.ReadCloser, error)
}

type ManifestListOptions struct {
	Manifests []*image.Info
	Platforms []string
}

type GetRepoImageOptions struct {
	IsImageIndex bool
}

type CopyImageOptions struct{}
