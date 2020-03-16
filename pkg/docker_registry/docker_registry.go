package docker_registry

import (
	"github.com/flant/werf/pkg/image"
)

type DockerRegistry interface {
	Tags(reference string) ([]string, error)
	GetRepoImage(reference string) (*image.Info, error)
	GetRepoImageList(reference string) ([]*image.Info, error)
	SelectRepoImageList(reference string, f func(*image.Info) bool) ([]*image.Info, error)
	DeleteRepoImage(repoImageList ...*image.Info) error
}

func NewDockerRegistry(_ string) (DockerRegistry, error) {
	return &Default{}, nil
}
