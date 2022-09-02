package bundles

import (
	"helm.sh/helm/v3/pkg/chart"

	bundles_registry "github.com/werf/werf/pkg/deploy/bundles/registry"
)

type BundlesRegistryClient interface {
	PullChartToCache(ref *bundles_registry.Reference) error
	LoadChart(ref *bundles_registry.Reference) (*chart.Chart, error)
	SaveChart(ch *chart.Chart, ref *bundles_registry.Reference) error
	PushChart(ref *bundles_registry.Reference) error
}
