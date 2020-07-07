package storage

import (
	"fmt"
	"sync"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/werf/pkg/werf/locker_with_retry"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/werf"
)

func NewKubernetesLockManager(namespace string, kubeClient kubernetes.Interface, kubeDynamicClient dynamic.Interface) *KuberntesLockManager {
	return &KuberntesLockManager{
		KubeClient:        kubeClient,
		KubeDynamicClient: kubeDynamicClient,
		Namespace:         namespace,
		LockerPerProject:  make(map[string]lockgate.Locker),
	}
}

type KuberntesLockManager struct {
	KubeClient        kubernetes.Interface
	KubeDynamicClient dynamic.Interface
	Namespace         string
	LockerPerProject  map[string]lockgate.Locker

	mux sync.Mutex
}

func (manager *KuberntesLockManager) getLockerForProject(projectName string) (lockgate.Locker, error) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if locker, hasKey := manager.LockerPerProject[projectName]; hasKey {
		return locker, nil
	}

	name := configMapName(projectName)
	if _, err := getOrCreateConfigMapWithNamespaceIfNotExists(manager.KubeClient, manager.Namespace, name); err != nil {
		return nil, err
	}

	locker := lockgate.NewKubernetesLocker(
		manager.KubeDynamicClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, name, manager.Namespace,
	)

	lockerWithRetry := locker_with_retry.NewLockerWithRetry(locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: 10, MaxReleaseAttempts: 10})

	manager.LockerPerProject[projectName] = lockerWithRetry

	return locker, nil
}

func (manager *KuberntesLockManager) LockStage(projectName, signature string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesStageLockName(projectName, signature), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) LockStageCache(projectName, signature string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesStageCacheLockName(projectName, signature), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) LockImage(projectName, imageName string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kuberntesImageLockName(projectName, imageName), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) LockStagesAndImages(projectName string, opts LockStagesAndImagesOptions) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kuberntesStagesAndImagesLockName(projectName), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{Shared: opts.GetOrCreateImagesOnly}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) LockDeployProcess(projectName string, releaseName string, kubeContextName string) (LockHandle, error) {
	if locker, err := manager.getLockerForProject(projectName); err != nil {
		return LockHandle{}, err
	} else {
		_, lock, err := locker.Acquire(kubernetesDeployReleaseLockName(projectName, releaseName, kubeContextName), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
		return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
	}
}

func (manager *KuberntesLockManager) Unlock(lock LockHandle) error {
	if locker, err := manager.getLockerForProject(lock.ProjectName); err != nil {
		return err
	} else {
		err := locker.Release(lock.LockgateHandle)
		if err != nil {
			logboek.ErrF("ERROR: unable to release lock for %q: %s\n", lock.LockgateHandle.LockName, err)
		}
		return err
	}
}

func kubernetesStageLockName(_, signature string) string {
	return fmt.Sprintf("stage/%s", signature)
}

func kubernetesStageCacheLockName(_, signature string) string {
	return fmt.Sprintf("stage-cache/%s", signature)
}

func kuberntesImageLockName(_, imageName string) string {
	return fmt.Sprintf("image/%s", imageName)
}

func kuberntesStagesAndImagesLockName(_ string) string {
	return fmt.Sprintf("stages_and_images")
}

func kubernetesDeployReleaseLockName(_ string, releaseName string, kubeContextName string) string {
	return fmt.Sprintf("release/%s;kube-context/%s", releaseName, kubeContextName)
}
