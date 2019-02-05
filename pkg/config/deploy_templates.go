package config

type DeployTemplates struct {
	HelmRelease     string
	HelmReleaseSlug bool
	Namespace       string
	NamespaceSlug   bool
}
