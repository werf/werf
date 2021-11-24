package bundles

import (
	"context"
	"fmt"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
)

func Publish(ctx context.Context, bundle *chart_extender.Bundle, bundleRef string, bundlesRegistryClient *registry.Client) error {
	r, err := registry.ParseReference(bundleRef)
	if err != nil {
		return fmt.Errorf("error parsing bundle ref %q: %s", bundleRef, err)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{}

	if err := logboek.Context(ctx).LogProcess("Saving bundle to the local chart helm cache").DoError(func() error {
		path, err := filepath.Abs(bundle.Dir)
		if err != nil {
			return err
		}

		ch, err := loader.Load(path)
		if err != nil {
			return fmt.Errorf("error loading chart %q: %s", path, err)
		}

		if err := bundlesRegistryClient.SaveChart(ch, r); err != nil {
			return fmt.Errorf("unable to save bundle to the local chart helm cache: %s", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %q", bundleRef).DoError(func() error {
		return bundlesRegistryClient.PushChart(r)
	}); err != nil {
		return err
	}

	return nil
}
