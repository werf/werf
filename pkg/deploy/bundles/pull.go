package bundles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	helmdownloader "github.com/werf/nelm/v2/pkg/helm/pkg/downloader"
	"github.com/werf/werf/v3/pkg/deploy/bundles/registry"
	"github.com/werf/werf/v3/pkg/ref"
)

const (
	ProvenanceStrategyVerify           = "verify"
	ProvenanceStrategyVerifyIfPossible = "verify-if-possible"
)

type PullOptions struct {
	HelmOptions nelmcommon.HelmOptions
	// ProvenanceKeyring is the path to the keyring with the public keys to verify the signature with.
	ProvenanceKeyring string
	// ProvenanceStrategy, if ProvenanceStrategyVerify, requires the bundle to have a valid signature
	// and, if ProvenanceStrategyVerifyIfPossible, requires a valid signature of a signed bundle.
	// Other strategies verify nothing.
	ProvenanceStrategy string
}

func Pull(ctx context.Context, bundleRef, destDir string, bundlesRegistryClient *registry.Client, opts PullOptions) error {
	r, err := ref.ParseReference(bundleRef)
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %q", bundleRef).DoError(func() error {
		return bundlesRegistryClient.PullChartToCache(ctx, r, opts.HelmOptions)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Exporting bundle %q", bundleRef).DoError(func() error {
		ch, err := bundlesRegistryClient.LoadChart(ctx, r, opts.HelmOptions)
		if err != nil {
			return fmt.Errorf("unable to load pulled chart: %w", err)
		}

		if err := verifyCachedChart(ctx, bundlesRegistryClient, r, ch, opts); err != nil {
			return fmt.Errorf("verify bundle %q signature: %w", bundleRef, err)
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

// verifyCachedChart verifies the cached chart archive against its provenance file as Helm verifies a
// downloaded one, since the exported chart directory can't be verified.
func verifyCachedChart(ctx context.Context, bundlesRegistryClient *registry.Client, r *ref.Reference, ch *chart.Chart, opts PullOptions) error {
	if opts.ProvenanceStrategy != ProvenanceStrategyVerify && opts.ProvenanceStrategy != ProvenanceStrategyVerifyIfPossible {
		return nil
	}

	chartArchive, provenance, err := bundlesRegistryClient.LoadChartArchive(ctx, r, opts.HelmOptions)
	if err != nil {
		return fmt.Errorf("load cached chart archive: %w", err)
	}

	if provenance == nil {
		if opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible {
			return nil
		}

		return fmt.Errorf("the bundle is not signed")
	}

	tmpDir, err := os.MkdirTemp("", "werf-bundle-verify-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// The provenance file refers to the archive by the name helm package gives it.
	archivePath := filepath.Join(tmpDir, fmt.Sprintf("%s-%s.tgz", ch.Name(), ch.Metadata.Version))
	if err := os.WriteFile(archivePath, chartArchive, 0o600); err != nil {
		return fmt.Errorf("write chart archive: %w", err)
	}

	if err := os.WriteFile(archivePath+".prov", provenance, 0o600); err != nil {
		return fmt.Errorf("write chart provenance file: %w", err)
	}

	if _, err := helmdownloader.VerifyChart(archivePath, archivePath+".prov", opts.ProvenanceKeyring); err != nil {
		return err
	}

	return nil
}
