package bundles

import (
	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	bundles_registry "github.com/werf/werf/v2/pkg/deploy/bundles/registry"
)

type BundlesRegistryClient interface {
	PullChartToCache(ref *bundles_registry.Reference, opts helmopts.HelmOptions) error
	LoadChart(ref *bundles_registry.Reference, opts helmopts.HelmOptions) (*chart.Chart, error)
	SaveChart(ch *chart.Chart, ref *bundles_registry.Reference, opts helmopts.HelmOptions) error
	PushChart(ref *bundles_registry.Reference, opts helmopts.HelmOptions) error
}
