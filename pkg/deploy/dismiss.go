package deploy

import (
	"fmt"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DismissOptions struct {
	WithNamespace bool
	KubeContext   string
}

func RunDismiss(release, namespace, kubeContext string, opts DismissOptions) error {
	if debug() {
		fmt.Printf("Dismiss options: %#v\n", opts)
		fmt.Printf("Namespace: %s\n", namespace)
	}

	err := helm.PurgeHelmRelease(release, helm.CommonHelmOptions{KubeContext: opts.KubeContext})
	if err != nil {
		return err
	}

	if opts.WithNamespace {
		logger.LogInfoF("Deleting kubernetes namespace '%s'\n", namespace)

		err := kube.Kubernetes.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete namespace %s: %s", namespace, err)
		}
	}

	return nil
}
