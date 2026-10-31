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
})

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
