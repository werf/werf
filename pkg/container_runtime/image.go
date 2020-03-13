package container_runtime

import "github.com/flant/werf/pkg/image"

type Image interface{}

type DockerImage struct {
	Image image.ImageInterface
}
