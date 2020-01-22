package render_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("helm render", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("base"), testDirPath)
	})

	It("should be rendered", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "render",
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
