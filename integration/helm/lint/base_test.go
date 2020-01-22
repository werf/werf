package lint_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("helm lint", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("base"), testDirPath)
	})

	It("should be linted", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "lint",
		)

		Ω(output).Should(ContainSubstring("1 chart(s) linted, no failures"))
	})
})
