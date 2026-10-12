package export

import (
	"context"
	"io"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Export label environment options", func() {
	ginkgo.BeforeEach(clearListEnv)
	ginkgo.DescribeTable("validates environment labels and excludes separator options", func(bare, suffix, invalid string) {
		ginkgo.GinkgoT().Setenv("WERF_EXPORT_ADD_LABEL", bare)
		ginkgo.GinkgoT().Setenv("WERF_EXPORT_ADD_LABEL_1", suffix)
		ginkgo.GinkgoT().Setenv("WERF_EXPORT_ADD_LABEL_SEPARATOR", ";")
		ginkgo.GinkgoT().Setenv("WERF_EXPORT_ADD_LABEL_SEPARATOR_1", "not-a-label")
		oldData := commonCmdData
		ginkgo.DeferCleanup(func() { commonCmdData = oldData })
		cmd := NewExportCmd(context.Background())
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		gomega.Expect(cmd.ParseFlags([]string{"--tag=example.com/image:tag", "--add-label=invalid-cli"})).To(gomega.Succeed())
		err := cmd.RunE(cmd, nil)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("unsupported --add-label value: invalid value \"" + invalid + "\"")))
	},
		ginkgo.Entry("bare label", "invalid-bare", "suffix=value", "invalid-bare"),
		ginkgo.Entry("suffixed label", "bare=value", "invalid-suffix", "invalid-suffix"),
		ginkgo.Entry("separator is not a label", "bare=value;other=value", "suffix=value", "invalid-cli"),
	)
})
