package docker_registry

import (
	"strings"

	"github.com/flant/werf/pkg/image"
)

type Default struct {
	*api
}

func NewDefault() *Default {
	d := &Default{}
	d.api = newAPI(false, false) // TODO
	return d
}

func (r *Default) GetRepoImageList(_ string) ([]*image.Info, error) {
	return nil, nil
}

func (r *Default) SelectRepoImageList(_ string, f func(*image.Info) bool) ([]*image.Info, error) {
	return nil, nil
}

func (r *Default) DeleteRepoImage(repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if err := r.deleteRepoImage(repoImage); err != nil {
			return err
		}
	}

	return nil
}

func (r *Default) deleteRepoImage(repoImage *image.Info) error {
	reference := strings.Join([]string{repoImage.Repository, repoImage.Digest}, "@")
	return r.api.deleteImageByReference(reference)
}
