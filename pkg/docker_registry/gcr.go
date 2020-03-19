package docker_registry

import (
	"strings"

	"github.com/flant/werf/pkg/image"
)

var GCRUrlPatterns = []string{"^container\\.cloud\\.google\\.com", "^gcr\\.io", "^.*\\.gcr\\.io"}

type GCR struct {
	*defaultImplementation
}

func (r *GCR) DeleteRepoImage(repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if err := r.deleteRepoImage(repoImage); err != nil {
			return err
		}
	}

	return nil
}

func (r *GCR) deleteRepoImage(repoImage *image.Info) error {
	reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
	return r.api.deleteImageByReference(reference)
}
