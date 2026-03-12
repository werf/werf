package bundles

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/pkg/common"
	chartv2util "github.com/werf/nelm/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/werf/v2/pkg/deploy/bundles/registry"
	"github.com/werf/werf/v2/pkg/ref"
)

func Pull(ctx context.Context, bundleRef, destDir string, bundlesRegistryClient *registry.Client, opts nelmcommon.HelmOptions) error {
	r, err := ref.ParseReference(bundleRef)
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %q", bundleRef).DoError(func() error {
		return bundlesRegistryClient.PullChartToCache(ctx, r, opts)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Exporting bundle %q", bundleRef).DoError(func() error {
		ch, err := bundlesRegistryClient.LoadChart(ctx, r, opts)
		if err != nil {
			return fmt.Errorf("unable to load pulled chart: %w", err)
		}

		if destDir == "" {
			err = chartv2util.SaveDir(ch, "")
			if err != nil {
				return err
			}
		} else {
			err = chartv2util.SaveIntoDir(ch, destDir)
			if err != nil {
				return fmt.Errorf("unable to save chart into local destination directory %q: %w", destDir, err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
