package kube_run

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Copy environment options", func() {
	ginkgo.BeforeEach(clearListEnv)
	ginkgo.DescribeTable("keeps environment and flag copy paths in order", func(name string, get func() []string) {
		oldData := cmdData
		ginkgo.DeferCleanup(func() { cmdData = oldData })
		cmdData.CopyFrom = []string{"/cli/one:/dest", "/cli/two:/dest"}
		cmdData.CopyTo = []string{"/cli/one:/dest", "/cli/two:/dest"}
		ginkgo.GinkgoT().Setenv(name, "/bare:/dest")
		ginkgo.GinkgoT().Setenv(name+"_1", "/suffix:/dest")
		gomega.Expect(get()).To(gomega.Equal([]string{"/bare:/dest", "/suffix:/dest", "/cli/one:/dest", "/cli/two:/dest"}))
	},
		ginkgo.Entry("copy from", "WERF_COPY_FROM", getCopyFromRaw),
		ginkgo.Entry("copy to", "WERF_COPY_TO", getCopyToRaw),
	)
})
