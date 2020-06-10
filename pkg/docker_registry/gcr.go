package docker_registry

import (
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
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gcr := &gcr{defaultImplementation: d}

	return gcr, nil
}

func (r *gcr) DeleteRepoImage(repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if err := r.deleteRepoImage(repoImage); err != nil {
			return err
		}
	}

	return nil
}

func (r *gcr) deleteRepoImage(repoImage *image.Info) error {
	reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
	return r.api.deleteImageByReference(reference)
}

func (r *gcr) String() string {
	return GcrImplementationName
}
