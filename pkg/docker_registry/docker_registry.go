package docker_registry

import (
	image2 "github.com/flant/werf/pkg/image"
)

type DockerRegistry interface {
	Tags(reference string) ([]string, error)
	List(reference string) ([]*image2.Info, error)
	Select(reference string, f func(*image2.Info) bool) ([]*image2.Info, error)
	DeleteImage(*image2.Info) error
}

/*
TODO: функция-фабрика
TODO: по общим параметрам определяет хранилище, выдает ошибку или конкретную реализацию
*/
func NewDockerRegistry(dockerRegistryAddress string) (DockerRegistry, error) {
	return nil, nil
}
