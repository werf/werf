package config

type DeployTemplates struct {
	HelmRelease             string
	HelmReleaseSlug         bool
	KubernetesNamespace     string
	KubernetesNamespaceSlug bool
}
