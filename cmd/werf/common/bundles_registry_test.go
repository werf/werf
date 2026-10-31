package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nelmcommon "github.com/werf/nelm/v2/pkg/common"
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
