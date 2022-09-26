package lock_manager

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/werf/pkg/kubeutils"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/locker_with_retry"
)

// NOTE: LockManager for not is not multithreaded due to the lack of support of contexts in the lockgate library
type LockManager struct {
	Namespace       string
	LockerWithRetry *locker_with_retry.LockerWithRetry
}

type ConfigMapLocker struct {
	ConfigMapName, Namespace string

	Locker lockgate.Locker
}

func NewConfigMapLocker(configMapName, namespace string, locker lockgate.Locker) *ConfigMapLocker {
	return &ConfigMapLocker{
		ConfigMapName: configMapName,
		Namespace:     namespace,
		Locker:        locker,
	}
}

func (locker *ConfigMapLocker) Acquire(lockName string, opts lockgate.AcquireOptions) (bool, lockgate.LockHandle, error) {
	if _, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(kube.Client, locker.Namespace, locker.ConfigMapName); err != nil {
		return false, lockgate.LockHandle{}, fmt.Errorf("unable to prepare kubernetes cm/%s in ns/%s: %w", locker.Namespace, locker.ConfigMapName, err)
	}
	return locker.Locker.Acquire(lockName, opts)
}

func (locker *ConfigMapLocker) Release(lock lockgate.LockHandle) error {
	return locker.Locker.Release(lock)
}

func NewLockManager(namespace string) (*LockManager, error) {
	configMapName := "werf-synchronization"

	locker := distributed_locker.NewKubernetesLocker(
		kube.DynamicClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, configMapName, namespace,
	)
	cmLocker := NewConfigMapLocker(configMapName, namespace, locker)
	lockerWithRetry := locker_with_retry.NewLockerWithRetry(context.Background(), cmLocker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: 10, MaxReleaseAttempts: 10})

	return &LockManager{
		Namespace:       namespace,
		LockerWithRetry: lockerWithRetry,
	}, nil
}

func (lockManager *LockManager) LockRelease(ctx context.Context, releaseName string) (lockgate.LockHandle, error) {
	// TODO: add support of context into lockgate
	lockManager.LockerWithRetry.Ctx = ctx
	_, handle, err := lockManager.LockerWithRetry.Acquire(fmt.Sprintf("release/%s", releaseName), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
	return handle, err
}

func (lockManager *LockManager) Unlock(handle lockgate.LockHandle) error {
	defer func() {
		lockManager.LockerWithRetry.Ctx = nil
	}()
	return lockManager.LockerWithRetry.Release(handle)
}
