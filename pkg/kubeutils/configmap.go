package kubeutils

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespaceIfNotExists(client kubernetes.Interface, namespace string) error {
	if _, err := client.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{}); errors.IsNotFound(err) {
		ns := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		if _, err := client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{}); errors.IsAlreadyExists(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("create Namespace %s error: %w", namespace, err)
		}
	} else if err != nil {
		return fmt.Errorf("get Namespace %s error: %w", namespace, err)
	}
	return nil
}

//  func createConfigMapIfNotExists(namespace, configMapName string) error {
//	  if _, err := client.CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{}); errors.IsNotFound(err) {
//		cm := &v1.ConfigMap{
//			ObjectMeta: metav1.ObjectMeta{Name: configMapName},
//		}
//
//		if _, err := client.CoreV1().ConfigMaps(namespace).Create(cm); errors.IsAlreadyExists(err) {
//			return nil
//		} else if err != nil {
//			return fmt.Errorf("create ConfigMap %s error: %w", configMapName, err)
//		}
//	  } else if err != nil {
//	  	  return fmt.Errorf("get ConfigMap %s error: %w", configMapName, err)
//	  }
//	  return nil
//  }

func GetOrCreateConfigMapWithNamespaceIfNotExists(client kubernetes.Interface, namespace, configMapName string) (*v1.ConfigMap, error) {
	obj, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	switch {
	case errors.IsNotFound(err):
		if err := CreateNamespaceIfNotExists(client, namespace); err != nil {
			return nil, err
		}

		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: configMapName},
		}

		obj, err := client.CoreV1().ConfigMaps(namespace).Create(context.Background(), cm, metav1.CreateOptions{})
		switch {
		case errors.IsAlreadyExists(err):
			obj, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("get ConfigMap %s error: %w", configMapName, err)
			}

			return obj, nil
		case err != nil:
			return nil, fmt.Errorf("create ConfigMap %s error: %w", cm.Name, err)
		default:
			return obj, nil
		}
	case err != nil:
		return nil, fmt.Errorf("get ConfigMap %s error: %w", configMapName, err)
	default:
		return obj, nil
	}
}
