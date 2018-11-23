package deploy

import (
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/kubedog/pkg/kube"
)

func getNamespace(namespaceOption string) string {
	if namespaceOption == "" {
		return kube.DefaultNamespace
	}
	return slug.Slug(namespaceOption)
}
