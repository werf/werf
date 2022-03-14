package docker_registry

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/werf/pkg/image"
)

type Interface interface {
	CreateRepo(ctx context.Context, reference string) error
	DeleteRepo(ctx context.Context, reference string) error
	Tags(ctx context.Context, reference string) ([]string, error)
	TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) error
	GetRepoImage(ctx context.Context, reference string) (*image.Info, error)
	TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error)
	IsRepoImageExists(ctx context.Context, reference string) (bool, error)
	DeleteRepoImage(ctx context.Context, repoImage *image.Info) error
	PushImage(ctx context.Context, reference string, opts *PushImageOptions) error
	MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateConfigFunc func(v1.Config) (v1.Config, error)) error

	String() string
}

type ApiInterface interface {
	GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error)
}
