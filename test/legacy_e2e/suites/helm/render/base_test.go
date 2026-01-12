package render_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("helm render", func() {
	BeforeEach(func(ctx SpecContext) {
		SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")
	})

	It("should be rendered", func(ctx SpecContext) {
		output := utils.SucceedCommandOutputString(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "render")

		for _, substrFormat := range []string{
			"# Source: %s/templates/010-secret.yaml",
			"# Source: %s/templates/020-backend.yaml",
			"# Source: %s/templates/090-frontend.yaml",
		} {
			Expect(output).Should(ContainSubstring(fmt.Sprintf(substrFormat, utils.ProjectName())))
		}
	})
})
