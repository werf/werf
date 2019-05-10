package deploy

import (
	"fmt"

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
	logProcessMsg := fmt.Sprintf("Running dismiss release %s", releaseName)
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess(logProcessMsg, logProcessOptions, func() error {
		return helm.PurgeHelmRelease(releaseName, namespace, opts.WithNamespace, opts.WithHooks)
	})
}
