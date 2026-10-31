package bundles

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	"github.com/werf/werf/v3/pkg/logging"
	"github.com/werf/werf/v3/pkg/ref"
)

var _ = Describe("Bundle addresses", func() {
	DescribeTable("should parse the source address with the tag encoded as Helm does",
		func(addr, expectedTag string) {
			parsed, err := ParseSourceAddr(addr)
			Expect(err).To(Succeed())
			Expect(parsed.RegistryAddress.Repo).To(Equal("registry.example.com/group/app"))
			Expect(parsed.RegistryAddress.Tag).To(Equal(expectedTag))
		},
		Entry("tag with build metadata", "registry.example.com/group/app:1.2.3+build.5", "1.2.3_build.5"),
		Entry("tag with the schema", "docker://registry.example.com/group/app:v1.2.3+build.5", "v1.2.3_build.5"),
		Entry("plain tag", "registry.example.com/group/app:latest", "latest"),
		Entry("no tag", "registry.example.com/group/app", "latest"),
	)

	DescribeTable("should parse the destination address with the tag as given",
		func(addr, expectedTag string) {
			parsed, err := ParseDestinationAddr(addr)
			Expect(err).To(Succeed())
			Expect(parsed.RegistryAddress.Repo).To(Equal("registry.example.com/group/app"))
			Expect(parsed.RegistryAddress.Tag).To(Equal(expectedTag))
		},
		Entry("tag with build metadata", "registry.example.com/group/app:1.2.3+build.5", "1.2.3+build.5"),
		Entry("plain tag", "registry.example.com/group/app:latest", "latest"),
		Entry("no tag", "registry.example.com/group/app", ""),
		Entry("no tag with the schema", "docker://registry.example.com/group/app", ""),
	)

	It("should parse an archive address", func() {
		for _, parse := range []func(string) (*ref.Addr, error){ParseSourceAddr, ParseDestinationAddr} {
			parsed, err := parse("archive:/tmp/bundle.tar.gz")
			Expect(err).To(Succeed())
			Expect(parsed.ArchiveAddress.Path).To(Equal("/tmp/bundle.tar.gz"))
		}
	})

	DescribeTable("should copy to a tag with build metadata",
		func(ctx0 context.Context, newFrom func(bundlesRegistryClient *BundlesRegistryClientStub, registryClient *DockerRegistryStub) BundleAccessor) {
			ctx := logging.WithLogger(ctx0)
			bundlesRegistryClient := NewBundlesRegistryClientStub()
			registryClient := NewDockerRegistryStub()

			toAddr, err := ParseDestinationAddr("registry2.example.com/group2/app:1.2.3+build.5")
			Expect(err).To(Succeed())

			Expect(newFrom(bundlesRegistryClient, registryClient).CopyTo(ctx, NewBundleAccessor(toAddr, BundleAccessorOptions{
				BundlesRegistryClient: bundlesRegistryClient,
				RegistryClient:        registryClient,
			}), copyToOptions{})).To(Succeed())

			Expect(bundlesRegistryClient.PushedRefs).To(Equal([]string{"registry2.example.com/group2/app:1.2.3_build.5"}))
			Expect(bundlesRegistryClient.StubCharts["registry2.example.com/group2/app:1.2.3_build.5"].Metadata.Version).To(Equal("1.2.3+build.5"))
		},
		Entry("from remote", func(bundlesRegistryClient *BundlesRegistryClientStub, registryClient *DockerRegistryStub) BundleAccessor {
			fromAddr, err := ParseSourceAddr("registry.example.com/group/app:1.0.0+build.1")
			Expect(err).To(Succeed())
			bundlesRegistryClient.StubCharts[fromAddr.RegistryAddress.FullName()] = newTestCopyChart()

			return NewBundleAccessor(fromAddr, BundleAccessorOptions{BundlesRegistryClient: bundlesRegistryClient, RegistryClient: registryClient})
		}),
		Entry("from archive", func(bundlesRegistryClient *BundlesRegistryClientStub, registryClient *DockerRegistryStub) BundleAccessor {
			return NewBundleArchive(NewBundleArchiveStubReader(newTestCopyChart(), nil), NewBundleArchiveStubWriter())
		}),
	)

	It("should pull a bundle by a tag with build metadata", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)
		repo := startTestRegistry() + "/group/app"

		Expect(Publish(ctx, writeTestChart(), repo, "1.2.3+build.5", newTestRegistryClient(ctx), PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})).To(Succeed())

		Expect(Pull(ctx, repo+":1.2.3+build.5", filepath.Join(GinkgoT().TempDir(), "app"), newTestRegistryClient(ctx), PullOptions{})).To(Succeed())
	})
})

func newTestCopyChart() *chart.Chart {
	return newWerfDataChart("apiVersion: v2\nname: app\nversion: 1.0.0+build.1\n", "")
}
