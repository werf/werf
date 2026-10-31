package bundles

import (
	"context"

	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	"github.com/werf/werf/v3/pkg/deploy/bundles/registry"
	bundles_registry "github.com/werf/werf/v3/pkg/ref"
)

var _ BundlesRegistryClient = (*registry.Client)(nil)

type BundlesRegistryClient interface {
	PullChartToCache(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error
	LoadChart(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) (*chart.Chart, error)
	SaveChartArchive(ctx context.Context, chartArchive []byte, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) error
	PushChart(ctx context.Context, ref *bundles_registry.Reference, tags []string, opts nelmcommon.HelmOptions) error
	TagCachedChart(ctx context.Context, from, to *bundles_registry.Reference, opts nelmcommon.HelmOptions) error
	IsChartSigned(ctx context.Context, ref *bundles_registry.Reference, opts nelmcommon.HelmOptions) (bool, error)
}
