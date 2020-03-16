package docker_registry

import (
	"strings"

	"github.com/flant/werf/pkg/image"
)

type Default struct {
	*api
}

func NewDefault(options APIOptions) *Default {
	d := &Default{}
	d.api = newAPI(options)
	return d
}

func (r *Default) GetRepoImageList(reference string) ([]*image.Info, error) {
	return r.SelectRepoImageList(reference, func(_ *image.Info) bool { return true })
}

func (r *Default) SelectRepoImageList(reference string, f func(*image.Info) bool) ([]*image.Info, error) {
	tags, err := r.api.Tags(reference)
	if err != nil {
		return nil, err
	}

	var repoImageList []*image.Info
	for _, tag := range tags {
		ref := strings.Join([]string{reference, tag}, ":")
		repoImage, err := r.GetRepoImage(ref)
		if err != nil {
			return nil, err
		}

		if f(repoImage) {
			repoImageList = append(repoImageList, repoImage)
		}
	}

	return repoImageList, nil
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
