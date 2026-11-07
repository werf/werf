package bundles

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/loader"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = Describe("Bundle download", func() {
	DescribeTable("should ask for the keyring to verify a signed bundle without one",
		func(ctx0 context.Context, strategy string) {
			ctx := logging.WithLogger(ctx0)
			repo := startTestRegistry() + "/group/app"
			publishSignedTestChart(ctx, repo)

			_, err := Download(ctx, repo, "1.0.0", GinkgoT().TempDir(), newTestRegistryClient(ctx), DownloadOptions{
				ProvenanceStrategy: strategy,
			})
			Expect(err).To(MatchError(ContainSubstring("pass --provenance-keyring")))
		},
		Entry("verify", ProvenanceStrategyVerify),
		Entry("verify if possible", "verify-if-possible"),
	)

	It("should download only the tag of the given name, not the version it matches as a constraint", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		Expect(Publish(ctx, writeTestChart(), repo, "1.0.0", PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})).To(Succeed())

		_, err := Download(ctx, repo, "1.x", GinkgoT().TempDir(), newTestRegistryClient(ctx), DownloadOptions{})
		Expect(err).To(MatchError(ContainSubstring("not found")))
	})

	It("should download without verifying a bundle of werf 3.12 whose version is not the version its tag names", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"
		client := newTestRegistryClient(ctx)
		_, err := client.Push(oldBundleArchive(repo, "0.0.0-x", nil), repo+":1.0.0_rc1", registry.PushOptStrictMode(false))
		Expect(err).To(Succeed())

		Expect(downloadTestChartVersion(ctx, repo, "1.0.0_rc1", DownloadOptions{ProvenanceStrategy: "never"})).To(Equal("0.0.0-x"))
	})
})

func downloadTestChartVersion(ctx context.Context, repo, tag string, opts DownloadOptions) string {
	chartPath, err := Download(ctx, repo, tag, filepath.Join(GinkgoT().TempDir(), "bundle"), newTestRegistryClient(ctx), opts)
	Expect(err).To(Succeed())

	ch, err := loader.LoadFile(ctx, chartPath)
	Expect(err).To(Succeed())

	return ch.Metadata.Version
}
