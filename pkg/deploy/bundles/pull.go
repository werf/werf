package bundles

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/logboek"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/helm/pkg/provenance"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
)

const (
	ProvenanceStrategyVerify           = "verify"
	ProvenanceStrategyVerifyIfPossible = "verify-if-possible"
)

type PullOptions struct {
	// ProvenanceKeyring is the path to the keyring with the public keys to verify the signature with.
	ProvenanceKeyring string
	// ProvenanceStrategy, if ProvenanceStrategyVerify, requires the bundle to have a valid signature
	// and, if ProvenanceStrategyVerifyIfPossible, requires a valid signature of a signed bundle.
	// Other strategies verify nothing.
	ProvenanceStrategy string
}

// Pull downloads the bundle pushed into repo under tag and unpacks it into destDir. The tag is
// literal, with "+" encoded as "_" as Helm does.
func Pull(ctx context.Context, repo, tag, destDir string, registryClient *registry.Client, opts PullOptions) error {
	verify := opts.ProvenanceStrategy == ProvenanceStrategyVerify || opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible

	bundleRef := fmt.Sprintf("%s:%s", repo, tag)

	var result *registry.PullResult
	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", bundleRef).DoError(func() error {
		var err error
		result, err = registryClient.PullContext(ctx, fmt.Sprintf("%s:%s", repo, encodeTag(tag)), registry.PullOptWithProv(verify), registry.PullOptIgnoreMissingProv(true))
		if err != nil {
			return fmt.Errorf("pull chart: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	if verify {
		if err := verifyChartArchive(ctx, bundleRef, result, opts); err != nil {
			return fmt.Errorf("verify bundle %s signature: %w", bundleRef, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	expandDir, err := os.MkdirTemp(filepath.Dir(destDir), "expand-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(expandDir)

	if err := chartv2util.Expand(expandDir, bytes.NewReader(result.Chart.Data)); err != nil {
		return fmt.Errorf("unpack chart: %w", err)
	}

	if err := os.Rename(filepath.Join(expandDir, result.Chart.Meta.Name), destDir); err != nil {
		return fmt.Errorf("move unpacked chart into %q: %w", destDir, err)
	}

	return nil
}

// verifyChartArchive verifies the chart archive against its provenance file with the name helm
// package gives the archive, which the provenance file refers to: the archive is pushed under tags
// that may differ from the chart version.
func verifyChartArchive(ctx context.Context, bundleRef string, result *registry.PullResult, opts PullOptions) error {
	if result.Prov.Digest == "" {
		if opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible {
			logboek.Context(ctx).Warn().LogF("Bundle %s is not signed, its signature is not verified\n", bundleRef)
			return nil
		}

		return errors.New("the bundle is not signed")
	}

	if opts.ProvenanceKeyring == "" {
		return fmt.Errorf("--provenance-strategy=%s needs the public keys to verify the signature with: pass --provenance-keyring", opts.ProvenanceStrategy)
	}

	signatory, err := provenance.NewFromKeyring(opts.ProvenanceKeyring, "")
	if err != nil {
		return fmt.Errorf("load keyring: %w", err)
	}

	if _, err := signatory.Verify(result.Chart.Data, result.Prov.Data, fmt.Sprintf("%s-%s.tgz", result.Chart.Meta.Name, result.Chart.Meta.Version)); err != nil {
		return err
	}

	logboek.Context(ctx).Default().LogF("Bundle %s signature verified\n", bundleRef)

	return nil
}
