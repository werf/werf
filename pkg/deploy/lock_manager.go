package deploy

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/werf/locker_with_retry"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/werf/pkg/kubeutils"
	"github.com/werf/werf/pkg/werf"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/lockgate"
)

type LockManager struct {
	Namespace string
	Locker    lockgate.Locker
}

func NewLockManager(ctx context.Context, namespace string) (*LockManager, error) {
	configMapName := "werf-synchronization"

	if _, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(kube.Client, namespace, configMapName); err != nil {
		return nil, err
	}

	locker := distributed_locker.NewKubernetesLocker(
		kube.DynamicClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, configMapName, namespace,
	)
	lockerWithRetry := locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: 10, MaxReleaseAttempts: 10})

	return &LockManager{
		Namespace: namespace,
		Locker:    lockerWithRetry,
	}, nil
}

func (lockManager *LockManager) LockRelease(ctx context.Context, releaseName string) (lockgate.LockHandle, error) {
	_, handle, err := lockManager.Locker.Acquire(fmt.Sprintf("release/%s", releaseName), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
	return handle, err
}

func (lockManager *LockManager) Unlock(handle lockgate.LockHandle) error {
	return lockManager.Locker.Release(handle)
}
