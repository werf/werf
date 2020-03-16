package docker_registry

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/flant/werf/pkg/image"
)

var GCRUrlPatterns = []string{"^container\\.cloud\\.google\\.com", "^gcr\\.io", "^.*\\.gcr\\.io"}

type GCR struct {
	*Default
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

func isGCR(reference string) (bool, error) {
	u, err := url.Parse(fmt.Sprintf("scheme://%s", reference))
	if err != nil {
		return false, err
	}

	for _, pattern := range GCRUrlPatterns {
		matched, err := regexp.MatchString(pattern, u.Host)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}
