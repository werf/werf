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
	"github.com/samber/lo"

	"github.com/werf/nelm/v2/pkg/action"
	nelmchart "github.com/werf/nelm/v2/pkg/chart"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/logging"
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

	DescribeTable("should push a long non-semver tag as is and under a chart version that fits a tag",
		func(ctx0 context.Context, tagLen int) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"
			tag := "feature_" + strings.Repeat("a", tagLen-len("feature_"))

			Expect(Publish(ctx, writeTestChart(), repo, tag, newTestRegistryClient(ctx), PublishOptions{
				ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
			})).To(Succeed())

			repository, err := name.NewRepository(repo, name.Insecure)
			Expect(err).To(Succeed())
			tags, err := remote.List(repository)
			Expect(err).To(Succeed())
			Expect(tags).To(HaveLen(2))
			Expect(tags).To(ContainElement(tag))

			versionTag := lo.Without(tags, tag)[0]
			Expect(len(versionTag)).To(BeNumerically("<=", 128))
			Expect(versionTag).To(HavePrefix("0.0.0-"))
			Expect(pullTestChartVersion(ctx, newTestRegistryClient(ctx), repo, versionTag)).To(Equal(versionTag))
			Expect(pullTestChartVersion(ctx, newTestRegistryClient(ctx), repo, tag)).To(Equal(versionTag))
		},
		Entry("111 characters", 111),
		Entry("112 characters", 112),
		Entry("128 characters", 128),
	)

	It("should publish a chart with Chart.yaml apiVersion v3 that renders with its publish-time values", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		chartDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v3\nname: app\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte("chartValue: chart\n"), 0o644)).To(Succeed())
		repo := startTestRegistry() + "/group/app"

		Expect(Publish(ctx, chartDir, repo, "1.0.0", newTestRegistryClient(ctx), PublishOptions{
			ChartPackOptions: action.ChartPackOptions{
				Destination: GinkgoT().TempDir(),
				ExtraValues: map[string]interface{}{"werf": map[string]interface{}{"repo": repo}},
			},
		})).To(Succeed())

		bundleDir := filepath.Join(GinkgoT().TempDir(), "bundle")
		Expect(Pull(ctx, repo, "1.0.0", bundleDir, newTestRegistryClient(ctx), PullOptions{})).To(Succeed())

		result, err := nelmchart.RenderChart(ctx, bundleDir, "release", "namespace", 1, nelmcommon.DeployTypeInstall, nil, nil, nelmchart.RenderChartOptions{
			HelmOptions: nelmcommon.HelmOptions{
				ChartLoadOpts: nelmcommon.ChartLoadOptions{ChartType: nelmcommon.LegacyChartTypeBundle},
			},
			LocalKubeVersion: "1.30.0",
		})
		Expect(err).To(Succeed())
		Expect(result.Values["Values"]).To(Equal(chartcommon.Values{
			"chartValue": "chart",
			"werf":       map[string]interface{}{"repo": repo},
		}))
	})

	DescribeTable("should verify the signature of a pulled bundle as requested",
		func(ctx0 context.Context, publish func(ctx context.Context, repo string), tag, strategy, expectedErr string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"
			publish(ctx, repo)

			err := Pull(ctx, repo, tag, filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
				ProvenanceKeyring:  testKeyringPath("helm-test-key.pub"),
				ProvenanceStrategy: strategy,
			})
			if expectedErr == "" {
				Expect(err).To(Succeed())
			} else {
				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
			}
		},
		Entry("signed, verify", publishSignedTestChart, "1.0.0", ProvenanceStrategyVerify, ""),
		Entry("signed, verify if possible", publishSignedTestChart, "1.0.0", ProvenanceStrategyVerifyIfPossible, ""),
		Entry("signed, verify, the tag other than the version", publishSignedTestChart, "v1.0.0", ProvenanceStrategyVerify, ""),
		Entry("tampered, verify", publishTamperedTestChart, "1.0.0", ProvenanceStrategyVerify, "sha256 sum does not match"),
		Entry("tampered, verify if possible", publishTamperedTestChart, "1.0.0", ProvenanceStrategyVerifyIfPossible, "sha256 sum does not match"),
		Entry("tampered, never", publishTamperedTestChart, "1.0.0", "never", ""),
		Entry("unsigned, verify", publishUnsignedTestChart, "1.0.0", ProvenanceStrategyVerify, "the bundle is not signed"),
		Entry("unsigned, verify if possible", publishUnsignedTestChart, "1.0.0", ProvenanceStrategyVerifyIfPossible, ""),
	)

	DescribeTable("should refuse to verify the signature",
		func(ctx0 context.Context, keyring, expectedErr string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"
			publishSignedTestChart(ctx, repo)

			err := Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
				ProvenanceKeyring:  keyring,
				ProvenanceStrategy: ProvenanceStrategyVerify,
			})
			Expect(err).To(MatchError(ContainSubstring(expectedErr)))
		},
		Entry("signed with another key", testKeyringPath("helm-password-key.secret"), "openpgp: signature made by unknown entity"),
		Entry("without a keyring", "", "pass --provenance-keyring"),
	)

	It("should deploy an unsigned bundle without a keyring if possible", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		publishUnsignedTestChart(ctx, repo)

		Expect(Pull(ctx, repo, "1.0.0", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{
			ProvenanceStrategy: ProvenanceStrategyVerifyIfPossible,
		})).To(Succeed())
	})

	It("should publish a signed chart with its signature", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		publishSignedTestChart(ctx, repo)

		result, err := newTestRegistryClient(ctx).Pull(repo+":1.0.0", registry.PullOptWithProv(true))
		Expect(err).To(Succeed())
		Expect(result.Prov.Data).NotTo(BeEmpty())
	})
})

func publishSignedTestChart(ctx context.Context, repo string) {
	Expect(Publish(ctx, writeTestChart(), repo, "v1.0.0", newTestRegistryClient(ctx), PublishOptions{
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

	_, err = newTestRegistryClient(ctx).Push(archive, repo+":1.0.0", registry.PushOptProvData(provenance))
	Expect(err).To(Succeed())
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
	client, err := registry.NewClient(registry.ClientOptCredentialsFile(testCredentialsPath()), registry.ClientOptPlainHTTP())
	Expect(err).To(Succeed())

	return client
}

func testCredentialsPath() string {
	return filepath.Join(GinkgoT().TempDir(), "config.json")
}

func writeTestChart() string {
	chartDir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: app\nversion: 0.1.0\n"), 0o644)).To(Succeed())

	return chartDir
}

func pullTestChartVersion(ctx context.Context, client *registry.Client, repo, tag string) string {
	result, err := client.Pull(repo + ":" + tag)
	Expect(err).To(Succeed())

	return result.Chart.Meta.Version
}
