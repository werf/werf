package storage

import (
	"github.com/flant/werf/pkg/docker_registry"
)

type RepoStagesStorage struct {
	StagesStorage  // FIXME
	DockerRegistry docker_registry.DockerRegistry
}

func NewRepoStagesStorage(dockerRegistry docker_registry.DockerRegistry) StagesStorage {
	return &RepoStagesStorage{DockerRegistry: dockerRegistry}
}

// TODO: Реализация интерфейса StagesStorage через низкоуровневый объект DockerRegistry
