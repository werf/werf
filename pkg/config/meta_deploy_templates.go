package config

type MetaDeployTemplates struct {
	HelmRelease     string
	HelmReleaseSlug bool
	Namespace       string
	NamespaceSlug   bool
}
