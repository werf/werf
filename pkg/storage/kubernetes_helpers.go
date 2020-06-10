package storage

import (
	"fmt"

	"github.com/werf/kubedog/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createNamespaceIfNotExists(namespace string) error {
	if _, err := kube.Kubernetes.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); errors.IsNotFound(err) {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}

		if _, err := kube.Kubernetes.CoreV1().Namespaces().Create(ns); errors.IsAlreadyExists(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("create Namespace %s error: %s", namespace, err)
		}
	} else if err != nil {
		return fmt.Errorf("get Namespace %s error: %s", namespace, err)
	}
	return nil
}

//func createConfigMapIfNotExists(namespace, configMapName string) error {
//	if _, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{}); errors.IsNotFound(err) {
//		cm := &v1.ConfigMap{
//			ObjectMeta: metav1.ObjectMeta{Name: configMapName},
//		}
//
//		if _, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Create(cm); errors.IsAlreadyExists(err) {
//			return nil
//		} else if err != nil {
//			return fmt.Errorf("create ConfigMap %s error: %s", configMapName, err)
//		}
//	} else if err != nil {
//		return fmt.Errorf("get ConfigMap %s error: %s", configMapName, err)
//	}
//	return nil
//}

func getOrCreateConfigMapWithNamespaceIfNotExists(namespace, configMapName string) (*v1.ConfigMap, error) {
	if obj, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{}); errors.IsNotFound(err) {
		if err := createNamespaceIfNotExists(namespace); err != nil {
			return nil, err
		}

		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: configMapName},
		}

		if obj, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Create(cm); errors.IsAlreadyExists(err) {
			if obj, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{}); err != nil {
				return nil, fmt.Errorf("get ConfigMap %s error: %s", configMapName, err)
			} else {
				return obj, err
			}
		} else if err != nil {
			return nil, fmt.Errorf("create ConfigMap %s error: %s", cm.Name, err)
		} else {
			return obj, nil
		}
	} else if err != nil {
		return nil, fmt.Errorf("get ConfigMap %s error: %s", configMapName, err)
	} else {
		return obj, nil
	}
}

func configMapName(projectName string) string {
	return fmt.Sprintf("werf-%s", projectName)
}
