package bundles

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/helm/pkg/chart/loader"
	v2chart "github.com/werf/nelm/pkg/helm/pkg/chart/v2"
	"github.com/werf/werf/v2/pkg/deploy/bundles/registry"
	"github.com/werf/werf/v2/pkg/ref"
)

type PublishOptions struct {
	HelmCompatibleChart bool
	RenameChart         string
	HelmOptions         nelmcommon.HelmOptions
}

func Publish(ctx context.Context, bundleDir, bundleRef string, bundlesRegistryClient *registry.Client, opts PublishOptions) error {
	r, err := ref.ParseReference(bundleRef)
	if err != nil {
		return fmt.Errorf("error parsing bundle ref %q: %w", bundleRef, err)
	}

	if err := logboek.Context(ctx).Default().LogProcess("Saving bundle to the local chart helm cache").DoError(func() error {
		path, err := filepath.Abs(bundleDir)
		if err != nil {
			return err
		}

		loadCtx := nelmcommon.ContextWithHelmOptions(ctx, opts.HelmOptions)
		ch, err := loader.Load(loadCtx, path)
		if err != nil {
			return fmt.Errorf("error loading chart %q: %w", path, err)
		}

		v2ch, ok := ch.(*v2chart.Chart)
		if !ok {
			return fmt.Errorf("unsupported chart type %T", ch)
		}

		if nameOverwrite := GetChartNameOverwrite(r.Repo, opts.RenameChart, opts.HelmCompatibleChart); nameOverwrite != nil {
			v2ch.Metadata.Name = *nameOverwrite
		}

		if err := bundlesRegistryClient.SaveChart(ctx, v2ch, r, opts.HelmOptions); err != nil {
			return fmt.Errorf("unable to save bundle to the local chart helm cache: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %q", bundleRef).DoError(func() error {
		return bundlesRegistryClient.PushChart(ctx, r, opts.HelmOptions)
	}); err != nil {
		return err
	}

	return nil
}
