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
}

func RunDismiss(releaseName, namespace, _ string, opts DismissOptions) error {
	if debug() {
		logger.LogServiceF("Dismiss options: %#v\n", opts)
		logger.LogServiceF("Namespace: %s\n", namespace)
	}

	logger.LogLn()
	logProcessMsg := fmt.Sprintf("Running dismiss release %s", releaseName)
	if err := logger.LogProcess(logProcessMsg, logger.LogProcessOptions{}, func() error {
		return helm.PurgeHelmRelease(releaseName)
	}); err != nil {
		return err
	}

	if opts.WithNamespace {
		logProcessMsg := fmt.Sprintf("Deleting kubernetes namespace %s", namespace)
		if err := logger.LogSecondaryProcessInline(logProcessMsg, func() error {
			return kube.Kubernetes.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		}); err != nil {
			return fmt.Errorf("failed to delete namespace %s: %s", namespace, err)
		}

		logger.OptionalLnModeOn()
	}

	return nil
}
