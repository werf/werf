package bundles

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/downloader"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/deploy/bundles/registry"
)

type PublishOptions struct {
	HelmCompatibleChart bool
	RenameChart         string
	HelmOptions         helmopts.HelmOptions
}

func Publish(ctx context.Context, bundleDir, bundleRef string, bundlesRegistryClient *registry.Client, opts PublishOptions) error {
	r, err := registry.ParseReference(bundleRef)
	if err != nil {
		return fmt.Errorf("error parsing bundle ref %q: %w", bundleRef, err)
	}

	if err := logboek.Context(ctx).Default().LogProcess("Saving bundle to the local chart helm cache").DoError(func() error {
		path, err := filepath.Abs(bundleDir)
		if err != nil {
			return err
		}

		ch, err := loader.Load(path, opts.HelmOptions)
		if err != nil {
			var e *downloader.ErrRepoNotFound
			if errors.As(err, &e) {
				return fmt.Errorf("%w. Please add the missing repos via 'helm repo add'", e)
			}

			return fmt.Errorf("error loading chart %q: %w", path, err)
		}

		if nameOverwrite := GetChartNameOverwrite(r.Repo, opts.RenameChart, opts.HelmCompatibleChart); nameOverwrite != nil {
			ch.Metadata.Name = *nameOverwrite
		}

		if err := bundlesRegistryClient.SaveChart(ctx, ch, r, opts.HelmOptions); err != nil {
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
