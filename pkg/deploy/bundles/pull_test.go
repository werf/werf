package bundles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/action"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = Describe("Bundle pull", func() {
	DescribeTable("should trust only the chart archive, not the chart metadata in the OCI config",
		func(ctx0 context.Context, config, strategy string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"

			packOpts := signingChartPackOptions()
			packOpts.Version = "1.0.0"
			packed, err := action.ChartPack(ctx, writeTestChart(), packOpts)
			Expect(err).To(Succeed())
			chartArchive, err := os.ReadFile(packed.ArchivePath)
			Expect(err).To(Succeed())
			provenanceData, err := os.ReadFile(packed.ProvenancePath)
			Expect(err).To(Succeed())
			pushRawChart(repo, "1.0.0", []byte(config), chartArchive, provenanceData)

			baseDir := GinkgoT().TempDir()
			victimFile := filepath.Join(baseDir, "victim", "file")
			Expect(os.MkdirAll(filepath.Dir(victimFile), 0o755)).To(Succeed())
			Expect(os.WriteFile(victimFile, []byte("victim"), 0o644)).To(Succeed())

			bundleDir := filepath.Join(baseDir, "bundle")
			Expect(Pull(ctx, repo, "1.0.0", bundleDir, newTestRegistryClient(ctx), PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: strategy,
			})).To(Succeed())

			Expect(filepath.Join(bundleDir, "Chart.yaml")).To(BeARegularFile())
			Expect(os.ReadFile(victimFile)).To(Equal([]byte("victim")))
		},
		Entry("a name escaping the directory, verify", `{"apiVersion":"v2","name":"../victim","version":"9.9.9"}`, ProvenanceStrategyVerify),
		Entry("a name escaping the directory, never verify", `{"apiVersion":"v2","name":"../victim","version":"9.9.9"}`, "never"),
		Entry("null, verify", `null`, ProvenanceStrategyVerify),
		Entry("null, never verify", `null`, "never"),
	)

	It("should pull the chart and its signature from the manifest the tag pointed to when the pull started", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		var tagManifestRequests atomic.Int32
		repo := startTestRegistryWith(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/manifests/1.0.0") {
					tagManifestRequests.Add(1)
				}

				next.ServeHTTP(w, r)
			})
		}) + "/group/app"
		publishSignedTestChart(ctx, repo)
		tagManifestRequests.Store(0)

		Expect(Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
			ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
			ProvenanceStrategy: ProvenanceStrategyVerify,
		})).To(Succeed())
		Expect(tagManifestRequests.Load()).To(BeEquivalentTo(1))
	})

	It("should stop pulling the chart when the context is done", func(ctx0 context.Context) {
		pullCtx, cancel := context.WithCancel(logging.WithLogger(ctx0))
		defer cancel()
		var isPulling atomic.Bool
		repo := startTestRegistryWith(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if isPulling.Load() && strings.Contains(r.URL.Path, "/blobs/") {
					cancel()
					time.Sleep(100 * time.Millisecond)
				}

				next.ServeHTTP(w, r)
			})
		}) + "/group/app"
		publishUnsignedTestChart(pullCtx, repo)
		isPulling.Store(true)

		err := Pull(pullCtx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(pullCtx), PullOptions{})
		Expect(err).To(MatchError(context.Canceled))
	})

	DescribeTable("should refuse a verified chart whose version is not the version its tag names",
		func(ctx0 context.Context, tag, strategy string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"

			packOpts := signingChartPackOptions()
			packOpts.Version = "1.0.0"
			packed, err := action.ChartPack(ctx, writeTestChart(), packOpts)
			Expect(err).To(Succeed())
			chartArchive, err := os.ReadFile(packed.ArchivePath)
			Expect(err).To(Succeed())
			provenanceData, err := os.ReadFile(packed.ProvenancePath)
			Expect(err).To(Succeed())
			pushRawChart(repo, encodeTag(tag), []byte(`{"apiVersion":"v2","name":"app","version":"1.0.0"}`), chartArchive, provenanceData)

			err = Pull(ctx, repo, tag, filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: strategy,
			})
			Expect(err).To(MatchError(ContainSubstring("has chart version \"1.0.0\", not the version its tag names")))
		},
		Entry("verify", "2.0.0", ProvenanceStrategyVerify),
		Entry("verify if possible", "2.0.0", ProvenanceStrategyVerifyIfPossible),
		Entry("the tag with the v prefix", "v2.0.0", ProvenanceStrategyVerify),
		Entry("the tag with build metadata", "1.0.0+1", ProvenanceStrategyVerify),
	)

	It("should pull without verifying a bundle of werf 3.12 whose version is not the version its tag names", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		client := newTestRegistryClient(ctx)
		_, err := client.Push(oldBundleArchive(repo, "0.0.0-x", nil), repo+":1.0.0_rc1", registry.PushOptStrictMode(false))
		Expect(err).To(Succeed())

		bundleDir := filepath.Join(GinkgoT().TempDir(), "app")
		Expect(Pull(ctx, repo, "1.0.0_rc1", bundleDir, client, PullOptions{ProvenanceStrategy: "never"})).To(Succeed())

		metadata, err := chartv2util.LoadChartfile(filepath.Join(bundleDir, "Chart.yaml"))
		Expect(err).To(Succeed())
		Expect(metadata.Version).To(Equal("0.0.0-x"))
	})

	It("should pull and verify a bundle from a repository with a long basename", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/" + strings.Repeat("a", 200)
		packOpts := signingChartPackOptions()
		packOpts.Version = "1.0.0"
		packed, err := action.ChartPack(ctx, writeTestChart(), packOpts)
		Expect(err).To(Succeed())
		chartArchive, err := os.ReadFile(packed.ArchivePath)
		Expect(err).To(Succeed())
		provenanceData, err := os.ReadFile(packed.ProvenancePath)
		Expect(err).To(Succeed())
		_, err = newTestRegistryClient(ctx).Push(chartArchive, repo+":1.0.0", registry.PushOptProvData(provenanceData), registry.PushOptStrictMode(false))
		Expect(err).To(Succeed())

		bundleDir := filepath.Join(GinkgoT().TempDir(), "app")
		Expect(Pull(ctx, repo, "1.0.0", bundleDir, newTestRegistryClient(ctx), PullOptions{
			ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
			ProvenanceStrategy: ProvenanceStrategyVerify,
		})).To(Succeed())
		Expect(filepath.Join(bundleDir, "Chart.yaml")).To(BeARegularFile())
	})

	It("should log the digest of the pulled bundle", func(ctx0 context.Context) {
		var output bytes.Buffer
		logger := logboek.NewLogger(&output, &output)
		logger.Streams().DisableLineWrapping()
		ctx := logboek.NewContext(ctx0, logger)
		repo := startTestRegistry() + "/group/app"
		publishUnsignedTestChart(ctx, repo)
		repository, err := newTestRegistryClient(ctx).Generic().Repository(repo)
		Expect(err).To(Succeed())
		manifest, err := repository.Resolve(ctx, "1.0.0")
		Expect(err).To(Succeed())

		Expect(Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{})).To(Succeed())
		Expect(output.String()).To(ContainSubstring(fmt.Sprintf("Pulling bundle %s:1.0.0 (digest %s)", repo, manifest.Digest)))
	})

	DescribeTable("should refuse to pull a tag that addresses a provenance file without reaching the registry",
		func(ctx0 context.Context, tag string) {
			ctx := logging.WithLogger(ctx0)
			var requests atomic.Int32
			repo := startTestRegistryWith(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					next.ServeHTTP(w, r)
				})
			}) + "/group/app"

			err := Pull(ctx, repo, tag, filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{})
			Expect(err).To(MatchError(fmt.Sprintf("tag %q is not allowed: tags ending in \".prov\" address provenance files", tag)))
			Expect(requests.Load()).To(BeZero())
		},
		Entry("semver tag", "1.2.3-rc.prov"),
		Entry("non-semver tag", "latest.prov"),
	)

	It("should not pull a tag that does not exist as a version constraint", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		publishUnsignedTestChart(ctx, repo)

		err := Pull(ctx, repo, "1.x", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{})
		Expect(err).To(MatchError(ContainSubstring("not found")))
		Expect(strings.Count(err.Error(), "1.x")).To(Equal(1))
	})

	It("should ask for the keyring to verify if possible a signed bundle without one", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		publishSignedTestChart(ctx, repo)

		err := Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
			ProvenanceStrategy: ProvenanceStrategyVerifyIfPossible,
		})
		Expect(err).To(MatchError(ContainSubstring("is signed: --provenance-strategy=verify-if-possible needs the public keys to verify the signature with: pass --provenance-keyring")))
	})

	It("should report who signed a verified bundle", func(ctx0 context.Context) {
		var output bytes.Buffer
		logger := logboek.NewLogger(&output, &output)
		logger.Streams().DisableLineWrapping()
		ctx := logboek.NewContext(ctx0, logger)
		repo := startTestRegistry() + "/group/app"
		publishSignedTestChart(ctx, repo)

		Expect(Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
			ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
			ProvenanceStrategy: ProvenanceStrategyVerify,
		})).To(Succeed())
		Expect(output.String()).To(ContainSubstring(fmt.Sprintf("Bundle %s:1.0.0 signature verified (signed by Helm Testing (This key should only be used for testing. DO NOT TRUST.) <helm-testing@helm.sh>)", repo)))
	})
})

// pushRawChart pushes a chart with the OCI config as given, which the registry client would generate
// from the chart.
func pushRawChart(repo, tag string, config, chartArchive, provenanceData []byte) {
	host, repoPath, _ := strings.Cut(repo, "/")
	baseURL := fmt.Sprintf("http://%s/v2/%s", host, repoPath)

	pushBlob := func(data []byte, mediaType string) map[string]interface{} {
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(data))

		resp, err := http.Post(fmt.Sprintf("%s/blobs/uploads/?digest=%s", baseURL, digest), "application/octet-stream", bytes.NewReader(data))
		Expect(err).To(Succeed())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		return map[string]interface{}{"mediaType": mediaType, "digest": digest, "size": len(data)}
	}

	manifest, err := json.Marshal(map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config":        pushBlob(config, registry.ConfigMediaType),
		"layers": []interface{}{
			pushBlob(chartArchive, registry.ChartLayerMediaType),
			pushBlob(provenanceData, registry.ProvLayerMediaType),
		},
	})
	Expect(err).To(Succeed())

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/manifests/%s", baseURL, tag), bytes.NewReader(manifest))
	Expect(err).To(Succeed())
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	resp, err := http.DefaultClient.Do(req)
	Expect(err).To(Succeed())
	Expect(resp.Body.Close()).To(Succeed())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated))
}
