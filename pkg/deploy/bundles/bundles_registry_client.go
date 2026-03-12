package bundles

import (
	"context"

	"github.com/werf/nelm/pkg/export/helm/chart"
	"github.com/werf/nelm/pkg/export/helm/werf/helmopts"
	bundles_registry "github.com/werf/werf/v2/pkg/ref"
)

type BundlesRegistryClient interface {
	PullChartToCache(ctx context.Context, ref *bundles_registry.Reference, opts helmopts.HelmOptions) error
	LoadChart(ctx context.Context, ref *bundles_registry.Reference, opts helmopts.HelmOptions) (*chart.Chart, error)
	SaveChart(ctx context.Context, ch *chart.Chart, ref *bundles_registry.Reference, opts helmopts.HelmOptions) error
	PushChart(ctx context.Context, ref *bundles_registry.Reference, opts helmopts.HelmOptions) error
}
