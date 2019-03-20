package deploy

import (
	"fmt"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/docker_registry"
)

type ImageInfoGetterStub struct {
	Name       string
	ImageTag   string
	ImagesRepo string
}

func (d *ImageInfoGetterStub) IsNameless() bool {
	return d.Name == ""
}

func (d *ImageInfoGetterStub) GetName() string {
	return d.Name
}

func (d *ImageInfoGetterStub) GetImageName() string {
	if d.Name == "" {
		return fmt.Sprintf("%s:%s", d.ImagesRepo, d.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", d.ImagesRepo, d.Name, d.ImageTag)
}

func (d *ImageInfoGetterStub) GetImageId() (string, error) {
	return docker_registry.ImageId(d.GetImageName())
}

type ImageInfo struct {
	Config          *config.Image
	WithoutRegistry bool
	ImagesRepo      string
	Tag             string
}

func (d *ImageInfo) IsNameless() bool {
	return d.Config.Name == ""
}

func (d *ImageInfo) GetName() string {
	return d.Config.Name
}

func (d *ImageInfo) GetImageName() string {
	if d.Config.Name == "" {
		return fmt.Sprintf("%s:%s", d.ImagesRepo, d.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", d.ImagesRepo, d.Config.Name, d.Tag)
}

func (d *ImageInfo) GetImageId() (string, error) {
	if d.WithoutRegistry {
		return "", nil
	}

	imageName := d.GetImageName()

	res, err := docker_registry.ImageId(imageName)
	if err != nil {
		logboek.LogErrorF("WARNING: Getting image %s id failed: %s\n", imageName, err)
		return "", nil
	}

	return res, nil
}
