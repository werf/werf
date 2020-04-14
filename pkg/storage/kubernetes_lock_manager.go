package storage

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/flant/kubedog/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flant/lockgate"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/werf"
)

func NewKubernetesLockManager(namespace string) *KuberntesLockManager {
	return &KuberntesLockManager{
		Namespace:        namespace,
		LockerPerProject: make(map[string]lockgate.Locker),
	}
}

type KuberntesLockManager struct {
	Namespace        string
	LockerPerProject map[string]lockgate.Locker

	mux sync.Mutex
}

func (manager *KuberntesLockManager) getLockerForProject(projectName string) (lockgate.Locker, error) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if locker, hasKey := manager.LockerPerProject[projectName]; hasKey {
		return locker, nil
	}

	if err := createNamespaceIfNotExists(manager.Namespace); err != nil {
		return nil, err
	}
	configMapName := fmt.Sprintf("werf-%s", projectName)
	if err := createConfigMapIfNotExists(manager.Namespace, configMapName); err != nil {
		return nil, err
	}

	locker := lockgate.NewKubernetesLocker(
		kube.DynamicClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, configMapName, manager.Namespace,
	)
	manager.LockerPerProject[projectName] = locker

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

func (manager *KuberntesLockManager) Unlock(lock LockHandle) error {
	if locker, err := manager.getLockerForProject(lock.ProjectName); err != nil {
		return err
	} else {
		err := locker.Release(lock.LockgateHandle)
		if err != nil {
			logboek.ErrF("ERROR: unable to release lock for %q: %s", lock.LockgateHandle.LockName)
		}
		return err
	}
}

func kubernetesStageLockName(_, signature string) string {
	return fmt.Sprintf("%s", signature)
}

func kubernetesStageCacheLockName(_, signature string) string {
	return fmt.Sprintf("%s.cache", signature)
}

func kuberntesImageLockName(_, imageName string) string {
	return fmt.Sprintf("%s.image", imageName)
}

func kuberntesStagesAndImagesLockName(_ string) string {
	return fmt.Sprintf("stages_and_images")
}

func createNamespaceIfNotExists(namespace string) error {
	if _, err := kube.Kubernetes.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); errors.IsNotFound(err) {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}

		if _, err := kube.Kubernetes.CoreV1().Namespaces().Create(ns); errors.IsAlreadyExists(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("unable to create Namespace %s: %s", ns.Name, err)
		}
	} else if err != nil {
		return err
	}
	return nil
}

func createConfigMapIfNotExists(namespace, configMapName string) error {
	if _, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{}); errors.IsNotFound(err) {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: configMapName},
		}

		if _, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Create(cm); errors.IsAlreadyExists(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("unable to create ConfigMap %s: %s", cm.Name, err)
		}
	} else if err != nil {
		return err
	}
	return nil
}
