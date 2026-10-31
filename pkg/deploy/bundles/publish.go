package bundles

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
)

type PublishOptions struct {
	// ChartPackOptions are the options to pack the chart with. The chart name and version are set
	// from the repo and the tag.
	ChartPackOptions action.ChartPackOptions
}

// Publish packs the chart and pushes it into repo under the chart version derived from tag and under
// tag itself, both with "+" encoded as "_" as Helm does.
func Publish(ctx context.Context, chartDir, repo, tag string, registryClient *registry.Client, opts PublishOptions) error {
	sv, err := BundleTagToChartVersion(ctx, tag, time.Now())
	if err != nil {
		return fmt.Errorf("unable to set chart version from bundle tag %q: %w", tag, err)
	}

	packOpts := opts.ChartPackOptions
	packOpts.Name = path.Base(repo)
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

	var provenance []byte
	if packResult.ProvenancePath != "" {
		provenance, err = os.ReadFile(packResult.ProvenancePath)
		if err != nil {
			return fmt.Errorf("read chart provenance file: %w", err)
		}
	}

	tags := chartTags(packResult.ChartVersion, tag)

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s with tags %s", repo, strings.Join(tags, ", ")).DoError(func() error {
		if _, err := registryClient.PushContext(ctx, chartArchive, fmt.Sprintf("%s:%s", repo, packResult.ChartVersion), registry.PushOptProvData(provenance), registry.PushOptExtraTags(tags[1:])); err != nil {
			return fmt.Errorf("push chart: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
