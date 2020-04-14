package storage

import "github.com/flant/werf/pkg/image"

func NewKubernetesStagesStorageCache(namespace string) *KubernetesStagesStorageCache {
	return &KubernetesStagesStorageCache{Namespace: namespace}
}

type KubernetesStagesStorageCache struct {
	Namespace string
}

func (cache *KubernetesStagesStorageCache) GetAllStages(projectName string) (bool, []image.StageID, error) {
	panic("no")
}

func (cache *KubernetesStagesStorageCache) GetStagesBySignature(projectName, signature string) (bool, []image.StageID, error) {
	panic("no")
}

func (cache *KubernetesStagesStorageCache) StoreStagesBySignature(projectName, signature string, stages []image.StageID) error {
	panic("no")
}

func (cache *KubernetesStagesStorageCache) DeleteStagesBySignature(projectName, signature string) error {
	panic("no")
}
