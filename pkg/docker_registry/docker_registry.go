package docker_registry

import (
	"github.com/flant/werf/pkg/image"
)

type DockerRegistry interface {
	Tags(reference string) ([]string, error)
	GetRepoImage(reference string) (*image.Info, error)
	TryGetRepoImage(reference string) (*image.Info, error)
	IsRepoImageExists(reference string) (bool, error)
	GetRepoImageList(reference string) ([]*image.Info, error)
	SelectRepoImageList(reference string, f func(*image.Info) bool) ([]*image.Info, error)
	DeleteRepoImage(repoImageList ...*image.Info) error
}

type DockerRegistryOptions struct {
	InsecureRegistry      bool
	SkipTlsVerifyRegistry bool
}

func NewDockerRegistry(_ string, options DockerRegistryOptions) (DockerRegistry, error) {
	return newDefaultImplementation(defaultImplementationOptions{apiOptions{
		InsecureRegistry:      options.InsecureRegistry,
		SkipTlsVerifyRegistry: options.SkipTlsVerifyRegistry,
	}})
}
