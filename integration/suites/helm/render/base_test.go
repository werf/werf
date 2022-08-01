package render_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("helm render", func() {
	BeforeEach(func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")
	})

	It("should be rendered", func() {
		output := utils.SucceedCommandOutputString(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"render",
		)

		for _, substrFormat := range []string{
			"# Source: %s/templates/010-secret.yaml",
			"# Source: %s/templates/020-backend.yaml",
			"# Source: %s/templates/090-frontend.yaml",
		} {
			Ω(output).Should(ContainSubstring(fmt.Sprintf(substrFormat, utils.ProjectName())))
		}
	})
})
