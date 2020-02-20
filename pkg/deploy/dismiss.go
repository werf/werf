package deploy

import (
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/helm"
)

type DismissOptions struct {
	WithNamespace bool
	WithHooks     bool
}

func RunDismiss(releaseName, namespace, _ string, opts DismissOptions) error {
	if debug() {
		logboek.LogF("Dismiss options: %#v\n", opts)
		logboek.LogF("Namespace: %s\n", namespace)
	}

	logboek.LogLn()
	logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
	return logboek.Default.LogProcess("Running dismiss", logProcessOptions, func() error {
		return helm.PurgeHelmRelease(releaseName, namespace, opts.WithNamespace, opts.WithHooks)
	})
}
