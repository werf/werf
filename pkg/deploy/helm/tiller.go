package helm

import (
	"github.com/flant/kubedog/pkg/kube"

	helm_kube "k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/tiller/environment"
)

var (
	TillerReleaseServer *tiller.ReleaseServer
)

func InitTiller() error {
	env := environment.New()
	kubeClient := helm_kube.New(nil)

	cfgmaps := driver.NewConfigMaps(kubeClient.KubernetesClientSet().CoreV1().ConfigMaps("kube-system"))
	env.Releases = storage.Init(cfgmaps)

	// kubeClient

	svc := tiller.NewReleaseServer(env, kube.Kubernetes, false)
}
