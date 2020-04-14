package common

import (
	"fmt"

	"github.com/flant/werf/pkg/storage"
	"github.com/flant/werf/pkg/werf"
)

const (
	WerfSynchronizationKubernetesNamespace = "werf-synchronization"
)

func GetStagesStorageCache(synchronization string) (storage.StagesStorageCache, error) {
	switch synchronization {
	case storage.LocalStagesStorageAddress:
		return storage.NewFileStagesStorageCache(werf.GetStagesStorageCacheDir()), nil
	case storage.KubernetesStagesStorageAddress:
		return storage.NewFileStagesStorageCache(werf.GetStagesStorageCacheDir()), nil
		// TODO
		//return storage.NewKubernetesStagesStorageCache(WerfSynchronizationKubernetesNamespace), nil
	default:
		panic(fmt.Sprintf("unknown synchronization param %q", synchronization))
	}
}

func GetStorageLockManager(synchronization string) (storage.LockManager, error) {
	switch synchronization {
	case storage.LocalStagesStorageAddress:
		return storage.NewGenericLockManager(werf.GetHostLocker()), nil
	case storage.KubernetesStagesStorageAddress:
		return storage.NewKubernetesLockManager(WerfSynchronizationKubernetesNamespace), nil
	default:
		panic(fmt.Sprintf("unknown synchronization param %q", synchronization))
	}
}
