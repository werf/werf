package deploy

import (
	"fmt"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/helm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DismissOptions struct {
	WithNamespace bool
}

func RunDismiss(releaseName, namespace, _ string, opts DismissOptions) error {
	if debug() {
		logboek.LogServiceF("Dismiss options: %#v\n", opts)
		logboek.LogServiceF("Namespace: %s\n", namespace)
	}

	logboek.LogLn()
	logProcessMsg := fmt.Sprintf("Running dismiss release %s", releaseName)
	if err := logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		return helm.PurgeHelmRelease(releaseName)
	}); err != nil {
		return err
	}

	if opts.WithNamespace {
		logProcessMsg := fmt.Sprintf("Deleting kubernetes namespace %s", namespace)
		if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
			return kube.Kubernetes.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		}); err != nil {
			return fmt.Errorf("failed to delete namespace %s: %s", namespace, err)
		}

		logboek.OptionalLnModeOn()
	}

	return nil
}
