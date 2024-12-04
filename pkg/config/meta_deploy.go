package config

type MetaDeploy struct {
	HelmChartConfig MetaDeployHelmChartConfig
	HelmChartDir    *string
	HelmRelease     *string
	HelmReleaseSlug *bool
	Namespace       *string
	NamespaceSlug   *bool
}

type MetaDeployHelmChartConfig struct {
	AppVersion *string
}
