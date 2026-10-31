package bundles

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/helm/pkg/provenance"
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
	r, err := ref.ParseReference(encodeAddrTag(bundleRef))
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
// downloaded one, since the exported chart directory can't be verified. The strategies are the ones
// Nelm's parseVerificationStrategy accepts: "later" and "never" verify nothing.
func verifyCachedChart(ctx context.Context, bundlesRegistryClient *registry.Client, r *ref.Reference, ch *chart.Chart, opts PullOptions) error {
	if opts.ProvenanceStrategy != ProvenanceStrategyVerify && opts.ProvenanceStrategy != ProvenanceStrategyVerifyIfPossible {
		return nil
	}

	chartArchive, provenanceData, err := bundlesRegistryClient.LoadChartArchive(ctx, r, opts.HelmOptions)
	if err != nil {
		return fmt.Errorf("load cached chart archive: %w", err)
	}

	if provenanceData == nil {
		if opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible {
			logboek.Context(ctx).Warn().LogF("Bundle %s is not signed, its signature is not verified\n", r.FullName())
			return nil
		}

		return fmt.Errorf("the bundle is not signed")
	}

	if opts.ProvenanceKeyring == "" {
		return fmt.Errorf("--provenance-strategy=%s needs the public keys to verify the signature with: pass --provenance-keyring", opts.ProvenanceStrategy)
	}

	signatory, err := provenance.NewFromKeyring(opts.ProvenanceKeyring, "")
	if err != nil {
		return fmt.Errorf("load keyring: %w", err)
	}

	// The provenance file refers to the archive by the name helm package gives it.
	if _, err := signatory.Verify(chartArchive, provenanceData, fmt.Sprintf("%s-%s.tgz", ch.Name(), ch.Metadata.Version)); err != nil {
		return err
	}

	logboek.Context(ctx).Default().LogF("Bundle %s signature verified\n", r.FullName())

	return nil
}
