package common

import (
	"context"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	ggcrregistry "github.com/google/go-containerregistry/pkg/registry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/werf/v3/pkg/deploy/bundles"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = DescribeTable("GetBundlesRegistryConnectionOptions",
	func(insecure, skipTLSVerify bool, expected nelmcommon.ChartRepoConnectionOptions) {
		Expect(GetBundlesRegistryConnectionOptions(&CmdData{
			InsecureRegistry:      &insecure,
			SkipTlsVerifyRegistry: &skipTLSVerify,
		})).To(Equal(expected))
	},
	Entry("none", false, false, nelmcommon.ChartRepoConnectionOptions{}),
	Entry("insecure", true, false, nelmcommon.ChartRepoConnectionOptions{ChartRepoInsecure: true}),
	Entry("skip TLS verify", false, true, nelmcommon.ChartRepoConnectionOptions{ChartRepoSkipTLSVerify: true}),
	Entry("both", true, true, nelmcommon.ChartRepoConnectionOptions{ChartRepoInsecure: true, ChartRepoSkipTLSVerify: true}),
)

var _ = DescribeTable("NewBundlesRegistryClient should publish and pull bundles on a plain HTTP registry on a loopback host",
	func(ctx0 context.Context, insecure, skipTLSVerify bool) {
		ctx := logging.WithLogger(ctx0)
		server := httptest.NewServer(ggcrregistry.New(ggcrregistry.Logger(log.New(io.Discard, "", 0))))
		DeferCleanup(server.Close)
		repo := strings.TrimPrefix(server.URL, "http://") + "/app"

		client, err := NewBundlesRegistryClient(ctx, &CmdData{
			DockerConfig:          lo.ToPtr(GinkgoT().TempDir()),
			InsecureRegistry:      &insecure,
			LogDebug:              lo.ToPtr(false),
			SkipTlsVerifyRegistry: &skipTLSVerify,
		})
		Expect(err).To(Succeed())

		chartDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: app\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(bundles.Publish(ctx, chartDir, repo, "latest", client, bundles.PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})).To(Succeed())

		Expect(bundles.Pull(ctx, repo, "latest", filepath.Join(GinkgoT().TempDir(), "app"), client, bundles.PullOptions{})).To(Succeed())
	},
	Entry("no options", false, false),
	Entry("insecure", true, false),
	Entry("skip TLS verify", false, true),
	Entry("insecure and skip TLS verify", true, true),
)

var _ = DescribeTable("isLoopbackHost",
	func(host string, expected bool) {
		Expect(isLoopbackHost(host)).To(Equal(expected))
	},
	Entry("localhost with a port", "localhost:5000", true),
	Entry("localhost", "localhost", true),
	Entry("IPv4 loopback", "127.0.0.1:32768", true),
	Entry("another IPv4 loopback", "127.1.2.3", true),
	Entry("IPv6 loopback with a port", "[::1]:5000", true),
	Entry("IPv6 loopback", "::1", true),
	Entry("remote host", "registry.example.com", false),
	Entry("remote IP", "10.0.0.1:5000", false),
	Entry("Docker Hub", "registry-1.docker.io", false),
)
