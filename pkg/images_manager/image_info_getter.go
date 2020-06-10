package images_manager

import (
	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
)

type ImageInfoGetter interface {
	IsNameless() bool
	GetName() string
	GetImageName() string
	GetImageID() (string, error)
	GetImageDigest() (string, error)
	GetImageTag() string
}

type ImageInfo struct {
	Name            string
	Tag             string
	WithoutRegistry bool
	ImagesRepo      storage.ImagesRepo

	info *image.Info
}

func (d *ImageInfo) getOrCreateInfo() (*image.Info, error) {
	if d.info == nil {
		repoImage, err := d.ImagesRepo.GetRepoImage(d.Name, d.Tag)
		if err != nil {
			return nil, err
		}

		d.info = repoImage
	}

	return d.info, nil
}

func (d *ImageInfo) IsNameless() bool {
	return d.Name == ""
}

func (d *ImageInfo) GetName() string {
	return d.Name
}

func (d *ImageInfo) GetImageName() string {
	return d.ImagesRepo.ImageRepositoryNameWithTag(d.Name, d.Tag)
}

func (d *ImageInfo) GetImageID() (string, error) {
	if d.WithoutRegistry {
		return "", nil
	}

	repoImage, err := d.getOrCreateInfo()
	if err != nil {
		logboek.LogWarnF("WARNING: Getting image %s id failed: %s\n", d.GetImageName(), err)
		return "", nil
	}

	return repoImage.ID, nil
}

func (d *ImageInfo) GetImageDigest() (string, error) {
	if d.WithoutRegistry {
		return "", nil
	}

	repoImage, err := d.getOrCreateInfo()
	if err != nil {
		logboek.LogWarnF("WARNING: Getting image %s digest failed: %s\n", d.GetImageName(), err)
		return "", nil
	}

	return repoImage.RepoDigest, nil
}

func (d *ImageInfo) GetImageTag() string {
	return d.Tag
}
