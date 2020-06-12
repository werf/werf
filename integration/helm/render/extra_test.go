package render_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

var _ = Describe("helm render with extra annotations and labels", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("base"), testDirPath)
	})

	It("should be rendered with extra annotations", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "render", "--add-annotation=anno1=value1", "--add-annotation=anno2=value2",
		)

		Ω(strings.Count(output, `annotations:
    anno1: value1
    anno2: value2`)).Should(Equal(4))
	})

	It("should be rendered with extra labels", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "render", "--add-label=label1=value1", "--add-label=label2=value2",
		)

		Ω(strings.Count(output, `labels:
    label1: value1
    label2: value2`)).Should(Equal(4))
	})
})
