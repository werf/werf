package storage

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/kubeutils"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/locker_with_retry"
)

func NewKubernetesLockManager(namespace string, kubeClient kubernetes.Interface, kubeDynamicClient dynamic.Interface, getConfigMapNameFunc func(projectName string) string) *KubernetesLockManager {
	return &KubernetesLockManager{
		KubeClient:           kubeClient,
		KubeDynamicClient:    kubeDynamicClient,
		Namespace:            namespace,
		LockerPerProject:     make(map[string]lockgate.Locker),
		GetConfigMapNameFunc: getConfigMapNameFunc,
	}
}

type KubernetesLockManager struct {
	KubeClient           kubernetes.Interface
	KubeDynamicClient    dynamic.Interface
	Namespace            string
	LockerPerProject     map[string]lockgate.Locker
	GetConfigMapNameFunc func(projectName string) string

	mux sync.Mutex
}

func (manager *KubernetesLockManager) getLockerForProject(ctx context.Context, projectName string) (lockgate.Locker, error) {
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

func (manager *KubernetesLockManager) LockStage(ctx context.Context, projectName, digest string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(ctx, projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesStageLockName(projectName, digest), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KubernetesLockManager) LockStageCache(ctx context.Context, projectName, digest string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(ctx, projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesStageCacheLockName(projectName, digest), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KubernetesLockManager) Unlock(ctx context.Context, lock LockHandle) error {
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

func kubernetesStageLockName(projectName, digest string) string {
	return fmt.Sprintf("%s/stage/%s", projectName, digest)
}

func kubernetesStageCacheLockName(projectName, digest string) string {
	return fmt.Sprintf("%s/stage-cache/%s", projectName, digest)
}
