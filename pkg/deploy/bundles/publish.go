package bundles

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/werf/v3/pkg/deploy/bundles/registry"
	"github.com/werf/werf/v3/pkg/ref"
)

type PublishOptions struct {
	// ChartPackOptions are the options to pack the chart with. The chart name and version are set
	// from bundleRef.
	ChartPackOptions action.ChartPackOptions
	HelmOptions      nelmcommon.HelmOptions
}

func Publish(ctx context.Context, chartDir, bundleRef string, bundlesRegistryClient *registry.Client, opts PublishOptions) error {
	r, err := ref.ParseReference(bundleRef)
	if err != nil {
		return fmt.Errorf("error parsing bundle ref %q: %w", bundleRef, err)
	}

	sv, err := BundleTagToChartVersion(ctx, r.Tag, time.Now())
	if err != nil {
		return fmt.Errorf("unable to set chart version from bundle tag %q: %w", r.Tag, err)
	}

	packOpts := opts.ChartPackOptions
	packOpts.Name = chartNameFromRepo(r.Repo)
	packOpts.Version = sv.String()

	var packResult *action.ChartPackResult
	if err := logboek.Context(ctx).Default().LogProcess("Packing bundle").DoError(func() error {
		packResult, err = action.ChartPack(ctx, chartDir, packOpts)
		if err != nil {
			return fmt.Errorf("pack chart %q: %w", chartDir, err)
		}

		return nil
	}); err != nil {
		return err
	}

	chartArchive, err := os.ReadFile(packResult.ArchivePath)
	if err != nil {
		return fmt.Errorf("read chart archive: %w", err)
	}

	tags := chartTags(packResult.ChartVersion, r.Tag)
	r.Tag = tags[0]

	if err := logboek.Context(ctx).Default().LogProcess("Saving bundle to the local chart helm cache").DoError(func() error {
		if err := bundlesRegistryClient.SaveChartArchive(ctx, chartArchive, r, opts.HelmOptions); err != nil {
			return fmt.Errorf("unable to save bundle to the local chart helm cache: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %q", bundleRef).DoError(func() error {
		return bundlesRegistryClient.PushChart(ctx, r, tags, opts.HelmOptions)
	}); err != nil {
		return err
	}

	return nil
}
