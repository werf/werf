package docker_registry

import (
	"context"
	"strings"

	"github.com/werf/werf/pkg/image"
)

const GcrImplementationName = "gcr"

var gcrPatterns = []string{"^container\\.cloud\\.google\\.com", "^gcr\\.io", "^.*\\.gcr\\.io"}

type gcr struct {
	*defaultImplementation
}

type GcrOptions struct {
	defaultImplementationOptions
}

func newGcr(options GcrOptions) (*gcr, error) {
	d, err := newDefaultAPIForImplementation(GcrImplementationName, options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gcr := &gcr{defaultImplementation: d}

	return gcr, nil
}

func (r *gcr) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
	return r.api.deleteImageByReference(ctx, reference)
}

func (r *gcr) String() string {
	return GcrImplementationName
}
