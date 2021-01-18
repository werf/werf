package giterminism_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = XDescribe("werf-giterminism.yaml", func() {
	BeforeEach(gitInit)

	It("werf-giterminism.yaml must be committed", func() {
		output, err := utils.RunCommand(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"config", "render",
		)

		Ω(err).Should(HaveOccurred())
		Ω(output).Should(ContainSubstring("giterminism configuration file 'werf-giterminism.yaml' must be committed"))
	})
})
