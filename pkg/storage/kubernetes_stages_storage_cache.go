package storage

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/werf/pkg/kubeutils"

	"k8s.io/client-go/kubernetes"

	"github.com/werf/logboek"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/werf/werf/pkg/image"
	v1 "k8s.io/api/core/v1"
)

const (
	StagesStorageCacheConfigMapKey = "stagesStorageCache"
)

type KubernetesStagesStorageCacheOptions struct {
	GetConfigMapNameFunc func(projectName string) string
}

func NewKubernetesStagesStorageCache(namespace string, kubeClient kubernetes.Interface, getConfigMapNameFunc func(projectName string) string) *KubernetesStagesStorageCache {
	return &KubernetesStagesStorageCache{KubeClient: kubeClient, Namespace: namespace, GetConfigMapNameFunc: getConfigMapNameFunc}
}

type KubernetesStagesStorageCache struct {
	KubeClient           kubernetes.Interface
	Namespace            string
	GetConfigMapNameFunc func(projectName string) string
}

type KubernetesStagesStorageCacheData struct {
	StagesBySignature map[string][]image.StageID `json:"stagesBySignature"`
}

func (cache *KubernetesStagesStorageCache) String() string {
	return fmt.Sprintf("kubernetes ns/%s", cache.Namespace)
}

func (cache *KubernetesStagesStorageCache) extractCacheData(obj *v1.ConfigMap) (*KubernetesStagesStorageCacheData, error) {
	if data, hasKey := obj.Data[StagesStorageCacheConfigMapKey]; hasKey {
		var cacheData *KubernetesStagesStorageCacheData

		if err := json.Unmarshal([]byte(data), &cacheData); err != nil {
			logboek.Error().LogF("Error unmarshalling stages storage cache json in cm/%s by key %q: %s: will ignore cache\n", obj.Name, StagesStorageCacheConfigMapKey, err)
			return nil, nil
		}

		return cacheData, nil
	} else {
		return nil, nil
	}
}

func (cache *KubernetesStagesStorageCache) unsetCacheData(obj *v1.ConfigMap) {
	if obj.Data != nil {
		delete(obj.Data, StagesStorageCacheConfigMapKey)
	}
}

func (cache *KubernetesStagesStorageCache) setCacheData(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) {
	if data, err := json.Marshal(cacheData); err != nil {
		panic(fmt.Sprintf("cannot marshal data %#v into json: %s", cacheData, err))
	} else {
		if obj.Data == nil {
			obj.Data = make(map[string]string)
		}
		obj.Data[StagesStorageCacheConfigMapKey] = string(data)
	}
}

func (cache *KubernetesStagesStorageCache) getConfigMapName(projectName string) string {
	if cache.GetConfigMapNameFunc != nil {
		return cache.GetConfigMapNameFunc(projectName)
	} else {
		return fmt.Sprintf("werf-%s", projectName)
	}
}

func (cache *KubernetesStagesStorageCache) GetAllStages(projectName string) (bool, []image.StageID, error) {
	if obj, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(cache.KubeClient, cache.Namespace, cache.GetConfigMapNameFunc(projectName)); err != nil {
		return false, nil, err
	} else if cacheData, err := cache.extractCacheData(obj); err != nil {
		return false, nil, err
	} else if cacheData != nil {
		var res []image.StageID
		for _, stagesBySignature := range cacheData.StagesBySignature {
			res = append(res, stagesBySignature...)
		}
		return true, res, nil
	}
	return false, nil, nil
}

func (cache *KubernetesStagesStorageCache) DeleteAllStages(projectName string) error {
	return cache.changeCacheData(projectName, func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error {
		cache.unsetCacheData(obj)
		return nil
	})
}

func (cache *KubernetesStagesStorageCache) GetStagesBySignature(projectName, signature string) (bool, []image.StageID, error) {
	if obj, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(cache.KubeClient, cache.Namespace, cache.GetConfigMapNameFunc(projectName)); err != nil {
		return false, nil, err
	} else if cacheData, err := cache.extractCacheData(obj); err != nil {
		return false, nil, err
	} else if cacheData != nil {
		if stages, hasKey := cacheData.StagesBySignature[signature]; hasKey {
			return true, stages, nil
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (cache *KubernetesStagesStorageCache) StoreStagesBySignature(projectName, signature string, stages []image.StageID) error {
	return cache.changeCacheData(projectName, func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error {
		if cacheData == nil {
			cacheData = &KubernetesStagesStorageCacheData{
				StagesBySignature: make(map[string][]image.StageID),
			}
		}
		cacheData.StagesBySignature[signature] = stages
		cache.setCacheData(obj, cacheData)
		return nil
	})
}

func (cache *KubernetesStagesStorageCache) DeleteStagesBySignature(projectName, signature string) error {
	return cache.changeCacheData(projectName, func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error {
		if cacheData != nil {
			delete(cacheData.StagesBySignature, signature)
			cache.setCacheData(obj, cacheData)
		}
		return nil
	})
}

func (cache *KubernetesStagesStorageCache) changeCacheData(projectName string, changeFunc func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error) error {
RETRY_CHANGE:

	if obj, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(cache.KubeClient, cache.Namespace, cache.GetConfigMapNameFunc(projectName)); err != nil {
		return err
	} else if cacheData, err := cache.extractCacheData(obj); err != nil {
		return err
	} else if cacheData != nil {
		if err := changeFunc(obj, cacheData); err != nil {
			return err
		}

		if _, err := cache.KubeClient.CoreV1().ConfigMaps(cache.Namespace).Update(context.Background(), obj, metav1.UpdateOptions{}); errors.IsConflict(err) {
			goto RETRY_CHANGE
		} else if err != nil {
			return fmt.Errorf("update cm/%s error: %s", obj.Name, err)
		}
	} else {
		if err := changeFunc(obj, cacheData); err != nil {
			return err
		}

		if _, err := cache.KubeClient.CoreV1().ConfigMaps(cache.Namespace).Update(context.Background(), obj, metav1.UpdateOptions{}); errors.IsConflict(err) {
			goto RETRY_CHANGE
		} else if err != nil {
			return fmt.Errorf("update cm/%s error: %s", obj.Name, err)
		}
	}

	return nil
}
