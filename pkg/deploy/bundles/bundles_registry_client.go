package bundles

import (
	"context"

	nelmcommon "github.com/werf/nelm/pkg/common"
	chart "github.com/werf/nelm/pkg/helm/pkg/chart/v2"
	bundles_registry "github.com/werf/werf/v2/pkg/ref"
)

type BundlesRegistryClient interface {
	PullChartToCache(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error
	LoadChart(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) (*chart.Chart, error)
	SaveChart(ctx context.Context, ch *chart.Chart, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error
	PushChart(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error
}
