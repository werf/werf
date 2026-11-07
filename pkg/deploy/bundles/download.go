package bundles

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/werf/nelm/v2/pkg/chart"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/downloader"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
)

const ProvenanceStrategyVerify = "verify"

type DownloadOptions struct {
	ChartRepoConnectionOptions nelmcommon.ChartRepoConnectionOptions
	// ProvenanceKeyring is the path to the keyring with the public keys to verify the signature with.
	ProvenanceKeyring string
	// ProvenanceStrategy, if ProvenanceStrategyVerify, requires the bundle to have a valid signature
	// and, if "verify-if-possible", requires a valid signature of a signed bundle. Other strategies
	// verify nothing.
	ProvenanceStrategy string
}

// Download downloads the bundle pushed into repo under tag into destDir, verifying its signature as
// opts say, and returns the path of the chart archive. The tag is literal, with "+" encoded as "_"
// as Helm does.
func Download(ctx context.Context, repo, tag, destDir string, registryClient *registry.Client, opts DownloadOptions) (string, error) {
	repo, err := normalizeRepo(repo)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	bundleRef := fmt.Sprintf("%s://%s:%s", registry.OCIScheme, repo, encodeTag(tag))

	chartPath, err := chart.DownloadChart(ctx, bundleRef, registryClient, destDir, chart.DownloadChartOptions{
		ChartRepoConnectionOptions: opts.ChartRepoConnectionOptions,
		ChartProvenanceKeyring:     opts.ProvenanceKeyring,
		ChartProvenanceStrategy:    opts.ProvenanceStrategy,
		ChartVersionLiteral:        true,
	})
	if errors.Is(err, downloader.ErrKeyringRequired) {
		return "", fmt.Errorf("%w: pass --provenance-keyring", err)
	} else if err != nil {
		return "", err
	}

	return chartPath, nil
}
