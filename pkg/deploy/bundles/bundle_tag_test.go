package bundles

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BundleTagToSemver", func() {
	It("converts arbitrary tag to semver if tag not already semver", func() {
		ctx := context.Background()
		now := time.Now()

		{
			sv, err := BundleTagToChartVersion(ctx, "latest", now)
			Expect(err).To(BeNil())
			Expect(sv.String()).To(Equal(fmt.Sprintf("0.0.0-%d-latest", now.Unix())))
		}

		{
			sv, err := BundleTagToChartVersion(ctx, "my-branch/ABC", now)
			Expect(err).To(BeNil())
			Expect(sv.String()).To(Equal(fmt.Sprintf("0.0.0-%d-my-branch-abc", now.Unix())))
		}

		{
			sv, err := BundleTagToChartVersion(ctx, "1.24.425-prerelease11", now)
			Expect(err).To(BeNil())
			Expect(sv.String()).To(Equal("1.24.425-prerelease11"))
		}

		{
			sv, err := BundleTagToChartVersion(ctx, "0.2.10", now)
			Expect(err).To(BeNil())
			Expect(sv.String()).To(Equal("0.2.10"))
		}
	})
})
