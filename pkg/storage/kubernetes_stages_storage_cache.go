package storage

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/kubeutils"
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
	StagesByDigest map[string][]image.StageID `json:"stagesByDigest"`
}

func (cache *KubernetesStagesStorageCache) String() string {
	return fmt.Sprintf("kubernetes ns/%s", cache.Namespace)
}

func (cache *KubernetesStagesStorageCache) extractCacheData(ctx context.Context, obj *v1.ConfigMap) (*KubernetesStagesStorageCacheData, error) {
	if data, hasKey := obj.Data[StagesStorageCacheConfigMapKey]; hasKey {
		var cacheData *KubernetesStagesStorageCacheData

		if err := json.Unmarshal([]byte(data), &cacheData); err != nil {
			logboek.Context(ctx).Error().LogF("Error unmarshalling storage cache json in cm/%s by key %q: %s: will ignore cache\n", obj.Name, StagesStorageCacheConfigMapKey, err)
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

//nolint:unused
func (cache *KubernetesStagesStorageCache) getConfigMapName(projectName string) string {
	if cache.GetConfigMapNameFunc != nil {
		return cache.GetConfigMapNameFunc(projectName)
	} else {
		return fmt.Sprintf("werf-%s", projectName)
	}
}

func (cache *KubernetesStagesStorageCache) GetAllStages(ctx context.Context, projectName string) (bool, []image.StageID, error) {
	if obj, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(cache.KubeClient, cache.Namespace, cache.GetConfigMapNameFunc(projectName)); err != nil {
		return false, nil, err
	} else if cacheData, err := cache.extractCacheData(ctx, obj); err != nil {
		return false, nil, err
	} else if cacheData != nil {
		var res []image.StageID
		for _, stagesByDigest := range cacheData.StagesByDigest {
			res = append(res, stagesByDigest...)
		}
		return true, res, nil
	}
	return false, nil, nil
}

func (cache *KubernetesStagesStorageCache) DeleteAllStages(ctx context.Context, projectName string) error {
	return cache.changeCacheData(ctx, projectName, func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error {
		cache.unsetCacheData(obj)
		return nil
	})
}

func (cache *KubernetesStagesStorageCache) GetStagesByDigest(ctx context.Context, projectName, digest string) (bool, []image.StageID, error) {
	if obj, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(cache.KubeClient, cache.Namespace, cache.GetConfigMapNameFunc(projectName)); err != nil {
		return false, nil, err
	} else if cacheData, err := cache.extractCacheData(ctx, obj); err != nil {
		return false, nil, err
	} else if cacheData != nil {
		if stages, hasKey := cacheData.StagesByDigest[digest]; hasKey {
			return true, stages, nil
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (cache *KubernetesStagesStorageCache) StoreStagesByDigest(ctx context.Context, projectName, digest string, stages []image.StageID) error {
	return cache.changeCacheData(ctx, projectName, func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error {
		if cacheData == nil {
			cacheData = &KubernetesStagesStorageCacheData{
				StagesByDigest: make(map[string][]image.StageID),
			}
		}
		cacheData.StagesByDigest[digest] = stages
		cache.setCacheData(obj, cacheData)
		return nil
	})
}

func (cache *KubernetesStagesStorageCache) DeleteStagesByDigest(ctx context.Context, projectName, digest string) error {
	return cache.changeCacheData(ctx, projectName, func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error {
		if cacheData != nil {
			delete(cacheData.StagesByDigest, digest)
			cache.setCacheData(obj, cacheData)
		}
		return nil
	})
}

func (cache *KubernetesStagesStorageCache) changeCacheData(ctx context.Context, projectName string, changeFunc func(obj *v1.ConfigMap, cacheData *KubernetesStagesStorageCacheData) error) error {
RETRY_CHANGE:

	obj, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(cache.KubeClient, cache.Namespace, cache.GetConfigMapNameFunc(projectName))
	if err != nil {
		return err
	}

	cacheData, err := cache.extractCacheData(ctx, obj)
	if err != nil {
		return err
	}

	if err := changeFunc(obj, cacheData); err != nil {
		return err
	}

	_, err = cache.KubeClient.CoreV1().ConfigMaps(cache.Namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
	if err != nil {
		if errors.IsConflict(err) {
			goto RETRY_CHANGE
		}

		return fmt.Errorf("update cm/%s error: %w", obj.Name, err)
	}

	return nil
}
