package storage

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/werf/pkg/kubeutils"
	"github.com/werf/werf/pkg/werf/locker_with_retry"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/werf"
)

func NewKubernetesLockManager(namespace string, kubeClient kubernetes.Interface, kubeDynamicClient dynamic.Interface, getConfigMapNameFunc func(projectName string) string) *KuberntesLockManager {
	return &KuberntesLockManager{
		KubeClient:           kubeClient,
		KubeDynamicClient:    kubeDynamicClient,
		Namespace:            namespace,
		LockerPerProject:     make(map[string]lockgate.Locker),
		GetConfigMapNameFunc: getConfigMapNameFunc,
	}
}

type KuberntesLockManager struct {
	KubeClient           kubernetes.Interface
	KubeDynamicClient    dynamic.Interface
	Namespace            string
	LockerPerProject     map[string]lockgate.Locker
	GetConfigMapNameFunc func(projectName string) string

	mux sync.Mutex
}

func (manager *KuberntesLockManager) getLockerForProject(ctx context.Context, projectName string) (lockgate.Locker, error) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if locker, hasKey := manager.LockerPerProject[projectName]; hasKey {
		return locker, nil
	}

	name := manager.GetConfigMapNameFunc(projectName)
	if _, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(manager.KubeClient, manager.Namespace, name); err != nil {
		return nil, err
	}

	locker := distributed_locker.NewKubernetesLocker(
		manager.KubeDynamicClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, name, manager.Namespace,
	)
	lockerWithRetry := locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: 10, MaxReleaseAttempts: 10})

	manager.LockerPerProject[projectName] = lockerWithRetry

	return locker, nil
}

func (manager *KuberntesLockManager) LockStage(ctx context.Context, projectName, signature string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(ctx, projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesStageLockName(projectName, signature), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) LockStageCache(ctx context.Context, projectName, signature string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(ctx, projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesStageCacheLockName(projectName, signature), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) LockImage(ctx context.Context, projectName, imageName string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(ctx, projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kuberntesImageLockName(projectName, imageName), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) Unlock(ctx context.Context, lock LockHandle) error {
	if locker, err := manager.getLockerForProject(ctx, lock.ProjectName); err != nil {
		return err
	} else {
		err := locker.Release(lock.LockgateHandle)
		if err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to release lock for %q: %s\n", lock.LockgateHandle.LockName, err)
		}
		return err
	}
}

// FIXME: v1.2 use the same locks names as generic lock manager (include project name into lock name)
func kubernetesStageLockName(_, signature string) string {
	return fmt.Sprintf("stage/%s", signature)
}

func kubernetesStageCacheLockName(_, signature string) string {
	return fmt.Sprintf("stage-cache/%s", signature)
}

func kuberntesImageLockName(_, imageName string) string {
	return fmt.Sprintf("image/%s", imageName)
}
