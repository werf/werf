package config

type MetaDeploy struct {
	HelmChartDir    *string
	HelmRelease     *string
	HelmReleaseSlug *bool
	Namespace       *string
	NamespaceSlug   *bool
}
