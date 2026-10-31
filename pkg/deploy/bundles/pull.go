package bundles

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader/archive"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/helm/pkg/downloader"
	"github.com/werf/nelm/v2/pkg/helm/pkg/getter"
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

// Pull downloads the bundle pushed into repo under tag and unpacks it into destDir.
func Pull(ctx context.Context, repo, tag, destDir string, registryClient *registry.Client, opts PullOptions) error {
	verify := opts.ProvenanceStrategy == ProvenanceStrategyVerify || opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible

	bundleRef := fmt.Sprintf("%s:%s", repo, tag)

	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	downloadDir, err := os.MkdirTemp(filepath.Dir(destDir), "download-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(downloadDir)

	chartDownloader := &downloader.ChartDownloader{
		Out:            io.Discard,
		Verify:         lo.Ternary(verify, downloader.VerifyLater, downloader.VerifyNever),
		Getters:        getter.Getters(),
		Options:        []getter.Option{getter.WithRegistryClient(registryClient)},
		RegistryClient: registryClient,
		ContentCache:   filepath.Join(downloadDir, "cache"),
	}

	var archivePath string
	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", bundleRef).DoError(func() error {
		archivePath, _, err = chartDownloader.DownloadTo(registry.OCIScheme+"://"+repo, tag, downloadDir)
		if err != nil {
			return fmt.Errorf("download chart: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	chartArchive, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("read chart archive: %w", err)
	}

	metadata, err := chartArchiveMetadata(chartArchive)
	if err != nil {
		return err
	}

	if verify {
		if err := verifyChartArchive(ctx, bundleRef, chartArchive, archivePath+".prov", metadata, opts); err != nil {
			return fmt.Errorf("verify bundle %s signature: %w", bundleRef, err)
		}
	}

	expandDir := filepath.Join(downloadDir, "chart")
	if err := chartv2util.ExpandFile(expandDir, archivePath); err != nil {
		return fmt.Errorf("unpack chart: %w", err)
	}

	if err := os.Rename(filepath.Join(expandDir, metadata.Name), destDir); err != nil {
		return fmt.Errorf("move unpacked chart into %q: %w", destDir, err)
	}

	return nil
}

// verifyChartArchive verifies the chart archive against the provenance file at provPath with
// the name helm package gives the archive, which the provenance file refers to: the archive is
// downloaded under its tag, which may differ from the chart version.
func verifyChartArchive(ctx context.Context, bundleRef string, chartArchive []byte, provPath string, metadata *chartArchiveMetadataFields, opts PullOptions) error {
	provenanceData, err := os.ReadFile(provPath)
	if errors.Is(err, fs.ErrNotExist) {
		if opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible {
			logboek.Context(ctx).Warn().LogF("Bundle %s is not signed, its signature is not verified\n", bundleRef)
			return nil
		}

		return errors.New("the bundle is not signed")
	} else if err != nil {
		return fmt.Errorf("read provenance file: %w", err)
	}

	if opts.ProvenanceKeyring == "" {
		return fmt.Errorf("--provenance-strategy=%s needs the public keys to verify the signature with: pass --provenance-keyring", opts.ProvenanceStrategy)
	}

	if opts.ProvenanceKeyring == "" {
		return fmt.Errorf("--provenance-strategy=%s needs the public keys to verify the signature with: pass --provenance-keyring", opts.ProvenanceStrategy)
	}

	signatory, err := provenance.NewFromKeyring(opts.ProvenanceKeyring, "")
	if err != nil {
		return fmt.Errorf("load keyring: %w", err)
	}

	if _, err := signatory.Verify(chartArchive, provenanceData, fmt.Sprintf("%s-%s.tgz", metadata.Name, metadata.Version)); err != nil {
		return err
	}

	logboek.Context(ctx).Default().LogF("Bundle %s signature verified\n", bundleRef)

	return nil
}

type chartArchiveMetadataFields struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func chartArchiveMetadata(chartArchive []byte) (*chartArchiveMetadataFields, error) {
	files, err := archive.LoadArchiveFiles(bytes.NewReader(chartArchive))
	if err != nil {
		return nil, fmt.Errorf("read chart archive: %w", err)
	}

	chartYAML, found := lo.Find(files, func(f *archive.BufferedFile) bool { return f.Name == chartv2util.ChartfileName })
	if !found {
		return nil, fmt.Errorf("%s not found in chart archive", chartv2util.ChartfileName)
	}

	metadata := &chartArchiveMetadataFields{}
	if err := yaml.Unmarshal(chartYAML.Data, metadata); err != nil {
		return nil, fmt.Errorf("parse %s: %w", chartv2util.ChartfileName, err)
	}

	return metadata, nil
}
