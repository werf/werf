package bundles

import (
	"context"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	ggcrregistry "github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/werf/v3/pkg/deploy/bundles/registry"
	"github.com/werf/werf/v3/pkg/logging"
	"github.com/werf/werf/v3/pkg/ref"
)

var _ = Describe("Bundle publish", func() {
	DescribeTable("should push the chart under the version and the tag encoded as Helm does",
		func(ctx0 context.Context, tag, expectedVersion string, expectedTags []string) {
			ctx := logging.WithLogger(ctx0)
			registryHost := startTestRegistry()
			repo := registryHost + "/group/app"

			Expect(Publish(ctx, writeTestChart(), repo, tag, newTestRegistryClient(ctx), PublishOptions{
				ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
			})).To(Succeed())

			repository, err := name.NewRepository(repo, name.Insecure)
			Expect(err).To(Succeed())
			tags, err := remote.List(repository)
			Expect(err).To(Succeed())
			Expect(tags).To(ConsistOf(expectedTags))

			for _, tag := range expectedTags {
				Expect(pullTestChartVersion(ctx, newTestRegistryClient(ctx), repo, tag)).To(Equal(expectedVersion))
			}
		},
		Entry("semver tag with build metadata", "1.2.3+build.5", "1.2.3+build.5", []string{"1.2.3_build.5"}),
		Entry("semver tag with the v prefix and build metadata", "v1.2.3+build.5", "1.2.3+build.5", []string{"1.2.3_build.5", "v1.2.3_build.5"}),
		Entry("semver tag", "1.2.3", "1.2.3", []string{"1.2.3"}),
	)

	It("should refuse a chart with Chart.yaml apiVersion v3", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		chartDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v3\nname: app\nversion: 0.1.0\n"), 0o644)).To(Succeed())

		err := Publish(ctx, chartDir, startTestRegistry()+"/group/app", "1.0.0", newTestRegistryClient(ctx), PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})
		Expect(err).To(MatchError(ContainSubstring("werf bundle commands don't support Chart.yaml apiVersion v3 yet")))
	})

	DescribeTable("should verify the signature of a pulled bundle as requested",
		func(ctx0 context.Context, publish func(ctx context.Context, repo string), strategy, expectedErr string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"
			publish(ctx, repo)

			err := Pull(ctx, repo+":1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: strategy,
			})
			if expectedErr == "" {
				Expect(err).To(Succeed())
			} else {
				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
			}
		},
		Entry("signed, verify", publishSignedTestChart, ProvenanceStrategyVerify, ""),
		Entry("signed, verify if possible", publishSignedTestChart, ProvenanceStrategyVerifyIfPossible, ""),
		Entry("tampered, verify", publishTamperedTestChart, ProvenanceStrategyVerify, "sha256 sum does not match"),
		Entry("tampered, verify if possible", publishTamperedTestChart, ProvenanceStrategyVerifyIfPossible, "sha256 sum does not match"),
		Entry("tampered, never", publishTamperedTestChart, "never", ""),
		Entry("unsigned, verify", publishUnsignedTestChart, ProvenanceStrategyVerify, "the bundle is not signed"),
		Entry("unsigned, verify if possible", publishUnsignedTestChart, ProvenanceStrategyVerifyIfPossible, ""),
	)

	It("should publish a signed chart with its signature", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		publishSignedTestChart(ctx, repo)

		client := newTestRegistryClient(ctx)
		r, err := ref.ParseReference(repo + ":1.0.0")
		Expect(err).To(Succeed())
		Expect(client.PullChartToCache(ctx, r, nelmcommon.HelmOptions{})).To(Succeed())
		Expect(client.IsChartSigned(ctx, r, nelmcommon.HelmOptions{})).To(BeTrue())
	})
})

func publishSignedTestChart(ctx context.Context, repo string) {
	Expect(Publish(ctx, writeTestChart(), repo, "1.0.0", newTestRegistryClient(ctx), PublishOptions{
		ChartPackOptions: signingChartPackOptions(),
	})).To(Succeed())
}

func publishUnsignedTestChart(ctx context.Context, repo string) {
	Expect(Publish(ctx, writeTestChart(), repo, "1.0.0", newTestRegistryClient(ctx), PublishOptions{
		ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
	})).To(Succeed())
}

// publishTamperedTestChart publishes the archive of a chart with the signature of another one of
// the same name and version.
func publishTamperedTestChart(ctx context.Context, repo string) {
	packOpts := signingChartPackOptions()
	packOpts.Name = "app"
	packOpts.Version = "1.0.0"
	signed, err := action.ChartPack(ctx, writeTestChart(), packOpts)
	Expect(err).To(Succeed())
	provenance, err := os.ReadFile(signed.ProvenancePath)
	Expect(err).To(Succeed())

	tamperedChartDir := writeTestChart()
	Expect(os.WriteFile(filepath.Join(tamperedChartDir, "values.yaml"), []byte("tampered: true\n"), 0o644)).To(Succeed())
	tampered, err := action.ChartPack(ctx, tamperedChartDir, action.ChartPackOptions{Destination: GinkgoT().TempDir(), Name: "app", Version: "1.0.0"})
	Expect(err).To(Succeed())
	archive, err := os.ReadFile(tampered.ArchivePath)
	Expect(err).To(Succeed())

	r, err := ref.ParseReference(repo + ":1.0.0")
	Expect(err).To(Succeed())
	client := newTestRegistryClient(ctx)
	Expect(client.SaveChartArchive(ctx, archive, provenance, r, nelmcommon.HelmOptions{})).To(Succeed())
	Expect(client.PushChart(ctx, r, []string{r.Tag}, nelmcommon.HelmOptions{})).To(Succeed())
}

func signingChartPackOptions() action.ChartPackOptions {
	return action.ChartPackOptions{
		Destination: GinkgoT().TempDir(),
		Sign:        true,
		SignKey:     "helm-test",
		SignKeyring: testKeyringPath("helm-test-key.secret"),
	}
}

func testKeyringPath(name string) string {
	keyringPath, err := filepath.Abs(filepath.Join("testdata", name))
	Expect(err).To(Succeed())

	return keyringPath
}

func startTestRegistry() string {
	server := httptest.NewServer(ggcrregistry.New(ggcrregistry.Logger(log.New(io.Discard, "", 0))))
	DeferCleanup(server.Close)

	return strings.TrimPrefix(server.URL, "http://")
}

func newTestRegistryClient(ctx context.Context) *registry.Client {
	tmpDir := GinkgoT().TempDir()

	cache, err := registry.NewCache(registry.CacheOptRoot(filepath.Join(tmpDir, "cache")))
	Expect(err).To(Succeed())

	client, err := registry.NewClient(ctx, registry.ClientOptCredentialsFile(filepath.Join(tmpDir, "config.json")), registry.ClientOptInsecure(true), registry.ClientOptCache(cache))
	Expect(err).To(Succeed())

	return client
}

func writeTestChart() string {
	chartDir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: app\nversion: 0.1.0\n"), 0o644)).To(Succeed())

	return chartDir
}

func pullTestChartVersion(ctx context.Context, client *registry.Client, repo, tag string) string {
	r, err := ref.ParseReference(repo + ":" + tag)
	Expect(err).To(Succeed())

	Expect(client.PullChartToCache(ctx, r, nelmcommon.HelmOptions{})).To(Succeed())

	ch, err := client.LoadChart(ctx, r, nelmcommon.HelmOptions{})
	Expect(err).To(Succeed())

	return ch.Metadata.Version
}
