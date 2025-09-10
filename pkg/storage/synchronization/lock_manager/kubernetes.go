package lock_manager

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	lockerPkg "github.com/werf/common-go/pkg/locker"
	"github.com/werf/common-go/pkg/locker_with_retry"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/kubeutils"
)

func NewKubernetes(
	namespace string,
	kubeClient kubernetes.Interface,
	kubeDynamicClient dynamic.Interface,
	getConfigMapNameFunc func(projectName string) string,
) *Kubernetes {
	return &Kubernetes{
		KubeClient:           kubeClient,
		KubeDynamicClient:    kubeDynamicClient,
		Namespace:            namespace,
		LockerPerProject:     make(map[string]lockgate.Locker),
		GetConfigMapNameFunc: getConfigMapNameFunc,
	}
}

type Kubernetes struct {
	KubeClient           kubernetes.Interface
	KubeDynamicClient    dynamic.Interface
	Namespace            string
	LockerPerProject     map[string]lockgate.Locker
	GetConfigMapNameFunc func(projectName string) string

	mux sync.Mutex
}

func (manager *Kubernetes) getLockerForProject(
	ctx context.Context,
	projectName string,
) (lockgate.Locker, error) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if locker, hasKey := manager.LockerPerProject[projectName]; hasKey {
		return locker, nil
	}

	name := manager.GetConfigMapNameFunc(projectName)
	if _, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(ctx, manager.KubeClient, manager.Namespace, name, true); err != nil {
		return nil, err
	}

	locker := distributed_locker.NewKubernetesLocker(
		manager.KubeDynamicClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, name, manager.Namespace,
	)
	lockerWithRetry := locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: maxAcquireAttempts, MaxReleaseAttempts: maxReleaseAttempts})

	manager.LockerPerProject[projectName] = lockerWithRetry

	return locker, nil
}

func (manager *Kubernetes) LockStage(
	ctx context.Context,
	projectName, digest string,
) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(ctx, projectName); err != nil {
		return LockHandle{}, err
	} else {
		options := lockerPkg.SetupDefaultOptions(ctx, lockgate.AcquireOptions{})
		_, lock, err := locker.Acquire(kubernetesStageLockName(projectName, digest), options)
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *Kubernetes) Unlock(ctx context.Context, lock LockHandle) error {
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
