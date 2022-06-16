package bundles

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/bundles/registry"
)

func Pull(ctx context.Context, bundleRef, destDir string, bundlesRegistryClient *registry.Client) error {
	r, err := registry.ParseReference(bundleRef)
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %q", bundleRef).DoError(func() error {
		return bundlesRegistryClient.PullChartToCache(r)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Exporting bundle %q", bundleRef).DoError(func() error {
		ch, err := bundlesRegistryClient.LoadChart(r)
		if err != nil {
			return fmt.Errorf("unable to load pulled chart: %w", err)
		}

		if destDir == "" {
			err = chartutil.SaveDir(ch, "")
			if err != nil {
				return err
			}
		} else {
			err = chartutil.SaveIntoDir(ch, destDir)
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
