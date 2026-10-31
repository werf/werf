package bundles

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/logging"
)

var _ = Describe("BundleTagToChartVersion", func() {
	now := time.Unix(1700000000, 0)

	DescribeTable("should use only a strict semver tag as the chart version",
		func(ctx context.Context, tag, expectedVersion string) {
			sv, err := BundleTagToChartVersion(logging.WithLogger(ctx), tag, now)
			Expect(err).NotTo(HaveOccurred())
			Expect(sv.String()).To(Equal(expectedVersion))
		},
		Entry("semver", "0.2.10", "0.2.10"),
		Entry("semver with a prerelease", "1.24.425-prerelease11", "1.24.425-prerelease11"),
		Entry("semver with a build", "1.2.3+b", "1.2.3+b"),
		Entry("semver with a prerelease and a build", "1.2.3-rc.1+build.5", "1.2.3-rc.1+build.5"),
		Entry("semver with the v prefix", "v1.2.3", "1.2.3"),
		Entry("semver with the v prefix and a prerelease", "v1.2.3-rc.1", "1.2.3-rc.1"),
		Entry("latest", "latest", fmt.Sprintf("0.0.0-%d-latest", now.Unix())),
		Entry("branch", "my-branch/ABC", fmt.Sprintf("0.0.0-%d-my-branch-abc", now.Unix())),
		Entry("commit SHA", "3f2a9c1", fmt.Sprintf("0.0.0-%d-3f2a9c1", now.Unix())),
		Entry("major and minor only", "1.2", fmt.Sprintf("0.0.0-%d-1-2", now.Unix())),
		Entry("major and minor only with the v prefix", "v1.2", fmt.Sprintf("0.0.0-%d-v1-2", now.Unix())),
		Entry("all digits", "20240101", fmt.Sprintf("0.0.0-%d-20240101", now.Unix())),
		Entry("leading zero", "01.2.3", fmt.Sprintf("0.0.0-%d-01-2-3", now.Unix())),
		Entry("capital V prefix", "V1.2.3", fmt.Sprintf("0.0.0-%d-v1-2-3", now.Unix())),
		Entry("branch name", "feature/Foo_1", fmt.Sprintf("0.0.0-%d-feature-foo-1", now.Unix())),
	)
})

var _ = DescribeTable("chartTags",
	func(chartVersion, tag string, expectedTags []string) {
		Expect(chartTags(chartVersion, tag)).To(Equal(expectedTags))
	},
	Entry("tag equal to the version", "1.2.3", "1.2.3", []string{"1.2.3"}),
	Entry("tag with the v prefix", "1.2.3", "v1.2.3", []string{"1.2.3", "v1.2.3"}),
	Entry("non-semver tag", "0.0.0-1700000000-latest", "latest", []string{"0.0.0-1700000000-latest", "latest"}),
	Entry("no tag", "1.2.3", "", []string{"1.2.3"}),
	Entry("version with build metadata", "1.2.3+build.5", "", []string{"1.2.3_build.5"}),
	Entry("tag equal to the encoded version", "1.2.3+build.5", "1.2.3_build.5", []string{"1.2.3_build.5"}),
	Entry("tag with build metadata", "1.2.3+build.5", "v1.2.3+build.5", []string{"1.2.3_build.5", "v1.2.3_build.5"}),
)
