package bundles

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/samber/lo"

	"github.com/werf/logboek"
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

// Pull downloads the bundle pushed into repo under tag and unpacks it into destDir. The tag is
// literal, with "+" encoded as "_" as Helm does. When verifying the signature, a semver tag must name
// the version of the chart.
func Pull(ctx context.Context, repo, tag, destDir string, registryClient *registry.Client, opts PullOptions) error {
	if err := checkTagIsNotProvenance(tag); err != nil {
		return err
	}

	repo, err := normalizeRepo(repo)
	if err != nil {
		return err
	}

	verify := downloader.VerifyNever
	switch opts.ProvenanceStrategy {
	case ProvenanceStrategyVerify:
		verify = downloader.VerifyAlways
	case ProvenanceStrategyVerifyIfPossible:
		verify = downloader.VerifyIfPossible
	}

	if verify == downloader.VerifyAlways && opts.ProvenanceKeyring == "" {
		return fmt.Errorf("--provenance-strategy=%s needs the public keys to verify the signature with: pass --provenance-keyring", opts.ProvenanceStrategy)
	}

	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	downloadDir, err := os.MkdirTemp(filepath.Dir(destDir), "download-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(downloadDir)

	bundleRef := fmt.Sprintf("%s:%s", repo, tag)

	repository, err := registryClient.Generic().Repository(repo)
	if err != nil {
		return err
	}

	// The chart and its provenance are pulled by digest so that both come from the same manifest
	// even if the tag moves meanwhile.
	manifest, err := repository.Resolve(ctx, encodeTag(tag))
	if err != nil {
		return fmt.Errorf("resolve bundle: %w", err)
	}

	var chartArchivePath string
	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s (digest %s)", bundleRef, manifest.Digest).DoError(func() error {
		chartDownloader := &downloader.ChartDownloader{
			Out:            logboek.Context(ctx).OutStream(),
			Verify:         verify,
			Keyring:        opts.ProvenanceKeyring,
			Getters:        getter.Getters(),
			Options:        []getter.Option{getter.WithRegistryClient(registryClient), getter.WithContext(ctx)},
			RegistryClient: registryClient,
			// DownloadTo refuses to run without a content cache. It reads the cache only by a
			// digest it gets for a versioned reference, never for a digest reference like ours,
			// and never writes it, yet a cache of its own keeps concurrent pulls apart regardless.
			ContentCache: filepath.Join(downloadDir, "cache"),
		}

		var verification *provenance.Verification
		chartArchivePath, verification, err = chartDownloader.DownloadTo(fmt.Sprintf("%s://%s@%s", registry.OCIScheme, repo, manifest.Digest), "", downloadDir)
		if err != nil {
			if opts.ProvenanceStrategy == ProvenanceStrategyVerifyIfPossible && opts.ProvenanceKeyring == "" && errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("bundle %s is signed: --provenance-strategy=%s needs the public keys to verify the signature with: pass --provenance-keyring", bundleRef, opts.ProvenanceStrategy)
			}

			return fmt.Errorf("pull and verify bundle %s: %w", bundleRef, err)
		}

		if verification != nil && verification.SignedBy != nil {
			signers := lo.Keys(verification.SignedBy.Identities)
			sort.Strings(signers)
			logboek.Context(ctx).Default().LogF("Bundle %s signature verified (signed by %s)\n", bundleRef, strings.Join(signers, ", "))
		}

		return nil
	}); err != nil {
		return err
	}

	chartArchive, err := os.Open(chartArchivePath)
	if err != nil {
		return fmt.Errorf("open chart archive: %w", err)
	}
	defer chartArchive.Close()

	expandDir, err := os.MkdirTemp(filepath.Dir(destDir), "expand-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(expandDir)

	if err := chartv2util.Expand(expandDir, chartArchive); err != nil {
		return fmt.Errorf("unpack chart: %w", err)
	}

	entries, err := os.ReadDir(expandDir)
	if err != nil {
		return fmt.Errorf("read unpacked chart: %w", err)
	}

	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("unpack chart: expected a single chart directory, got %d entries", len(entries))
	}

	// Nelm binds a semver tag to the chart version only when pulling by tag, not by digest. Unverified
	// pulls skip the binding: werf 3.12 and older published bundles under tags not naming their
	// version.
	if tagVersion, err := semver.StrictNewVersion(strings.ReplaceAll(strings.TrimPrefix(tag, "v"), "_", "+")); err == nil && verify != downloader.VerifyNever {
		metadata, err := chartv2util.LoadChartfile(filepath.Join(expandDir, entries[0].Name(), chartv2util.ChartfileName))
		if err != nil {
			return fmt.Errorf("load chart metadata: %w", err)
		}

		if metadata.Version != tagVersion.Original() {
			return fmt.Errorf("bundle %s has chart version %q, not the version its tag names", bundleRef, metadata.Version)
		}
	}

	if err := os.Rename(filepath.Join(expandDir, entries[0].Name()), destDir); err != nil {
		return fmt.Errorf("move unpacked chart into %q: %w", destDir, err)
	}

	return nil
}
