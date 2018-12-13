package deploy

import (
	"fmt"

	"github.com/flant/kubedog/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DismissOptions struct {
	WithNamespace bool
	KubeContext   string
}

func RunDismiss(releaseName, namespace, kubeContext string, opts DismissOptions) error {
	if debug() {
		fmt.Printf("Dismiss options: %#v\n", opts)
		fmt.Printf("Namespace: %s\n", namespace)
	}

	err := PurgeHelmRelease(releaseName, CommonHelmOptions{KubeContext: opts.KubeContext})
	if err != nil {
		return err
	}

	if opts.WithNamespace {
		fmt.Printf("# Deleting kubernetes namespace '%s'...\n", namespace)

		err := kube.Kubernetes.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete namespace %s: %s", namespace, err)
		}
	}

	return nil
}
