package history

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestHistory(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Release History Suite")
}

var _ = ginkgo.Describe("release history revisions limit", func() {
	ginkgo.DescribeTable("uses the environment default and lets the flag override it",
		func(env string, args []string, expected int) {
			ginkgo.GinkgoT().Setenv("WERF_RELEASE_HISTORY_REVISIONS_LIMIT", env)
			cmd := NewCmd(context.Background())
			gomega.Expect(cmd.ParseFlags(args)).To(gomega.Succeed())
			gomega.Expect(cmdData.RevisionsLimit).To(gomega.Equal(expected))
		},
		ginkgo.Entry("unset", "", []string{}, 0),
		ginkgo.Entry("environment", "5", []string{}, 5),
		ginkgo.Entry("flag overrides environment", "5", []string{"--revisions-limit=2"}, 2),
		ginkgo.Entry("explicit unlimited overrides environment", "5", []string{"--revisions-limit=0"}, 0),
	)

	ginkgo.It("rejects a malformed environment default", func() {
		ginkgo.GinkgoT().Setenv("WERF_RELEASE_HISTORY_REVISIONS_LIMIT", "invalid")
		gomega.Expect(func() { NewCmd(context.Background()) }).To(gomega.PanicWith(gomega.ContainSubstring("WERF_RELEASE_HISTORY_REVISIONS_LIMIT")))
	})
})
