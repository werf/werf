package render_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("helm render with extra annotations and labels", func() {
	BeforeEach(func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")
	})

	It("should be rendered with extra annotations (except hooks)", func() {
		output := utils.SucceedCommandOutputString(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"render", "--add-annotation=anno1=value1", "--add-annotation=anno2=value2",
		)

		Ω(strings.Count(output, `annotations:
    anno1: value1
    anno2: value2`)).Should(Equal(3))
	})

	It("should be rendered with extra labels (except hooks)", func() {
		output := utils.SucceedCommandOutputString(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"render", "--add-label=label1=value1", "--add-label=label2=value2",
		)

		Ω(strings.Count(output, `labels:
    label1: value1
    label2: value2`)).Should(Equal(3))
	})
})
