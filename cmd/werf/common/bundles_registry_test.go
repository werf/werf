package common

import (
	"bytes"
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

	"github.com/werf/logboek"
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

var _ = DescribeTable("NewBundlesRegistryClient should publish and download bundles on a plain HTTP registry on a loopback host unless asked for HTTPS",
	func(ctx0 context.Context, insecure, skipTLSVerify, expectedPlainHTTP bool) {
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
		Expect(bundles.Publish(ctx, chartDir, repo, "latest", bundles.PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})).To(Succeed())

		_, err = bundles.Download(ctx, repo, "latest", GinkgoT().TempDir(), client, bundles.DownloadOptions{})
		if expectedPlainHTTP {
			Expect(err).To(Succeed())
		} else {
			Expect(err).To(MatchError(ContainSubstring("http: server gave HTTP response to HTTPS client")))
		}
	},
	Entry("no options", false, false, true),
	Entry("insecure", true, false, true),
	Entry("skip TLS verify", false, true, false),
	Entry("insecure and skip TLS verify", true, true, true),
)

var _ = DescribeTable("NewBundlesRegistryClient should print the registry client output only with debug logging",
	func(ctx0 context.Context, logDebug bool) {
		var output bytes.Buffer
		ctx := logboek.NewContext(ctx0, logboek.NewLogger(&output, &output))
		server := httptest.NewServer(ggcrregistry.New(ggcrregistry.Logger(log.New(io.Discard, "", 0))))
		DeferCleanup(server.Close)
		repo := strings.TrimPrefix(server.URL, "http://") + "/app"

		client, err := NewBundlesRegistryClient(ctx, &CmdData{
			DockerConfig:          lo.ToPtr(GinkgoT().TempDir()),
			InsecureRegistry:      lo.ToPtr(false),
			LogDebug:              &logDebug,
			SkipTlsVerifyRegistry: lo.ToPtr(false),
		})
		Expect(err).To(Succeed())

		chartDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: app\nversion: 0.1.0\n"), 0o644)).To(Succeed())
		Expect(bundles.Publish(ctx, chartDir, repo, "latest", bundles.PublishOptions{
			ChartPackOptions: action.ChartPackOptions{Destination: GinkgoT().TempDir()},
		})).To(Succeed())
		_, err = bundles.Download(ctx, repo, "latest", GinkgoT().TempDir(), client, bundles.DownloadOptions{})
		Expect(err).To(Succeed())

		if logDebug {
			Expect(output.String()).To(ContainSubstring("Pulled: "))
		} else {
			Expect(output.String()).NotTo(ContainSubstring("Pulled: "))
		}
	},
	Entry("without debug logging", false),
	Entry("with debug logging", true),
)
