package giterminism_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config stapel", func() {
	BeforeEach(CommonBeforeEach)

	Context("git.branch", func() {
		type entry struct {
			allowStapelGitBranch bool
			expectedErrSubstring string
		}

		DescribeTable("config.stapel.git.allowBranch",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", `
image: test
from: alpine
git:
- url: https://github.com/werf/werf.git
  branch: test
  to: /app
`)
				gitAddAndCommit("werf.yaml")

				if e.allowStapelGitBranch {
					contentToAppend := `
config:
  stapel:
    git:
      allowBranch: true`
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
				}

				output, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"config", "render",
				)

				if e.expectedErrSubstring != "" {
					Ω(err).Should(HaveOccurred())
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(err).ShouldNot(HaveOccurred())
				}
			},
			Entry("the remote git branch not allowed", entry{
				expectedErrSubstring: "the configuration with external dependency found in the werf config: git branch directive not allowed",
			}),
			Entry("the remote git branch allowed", entry{
				allowStapelGitBranch: true,
			}),
		)
	})

})
