package giterminism_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config", func() {
	BeforeEach(func() {
		gitInit()
		utils.CopyIn(utils.FixturePath("config"), SuiteData.TestDirPath)
		gitAddAndCommit("werf-giterminism.yaml")
	})

	const minimalConfigContent = `
configVersion: 1
project: none
---
`

	Context("regular files", func() {
		type entry struct {
			allowUncommitted        bool
			addConfig               bool
			commitConfig            bool
			changeConfigAfterCommit bool
			expectedErrSubstring    string
		}

		DescribeTable("config.allowUncommitted",
			func(e entry) {
				var contentToAppend string
				if e.allowUncommitted {
					contentToAppend = `
config:
  allowUncommitted: true`
				} else {
					contentToAppend = `
config:
  allowUncommitted: false`
				}
				fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
				gitAddAndCommit("werf-giterminism.yaml")

				if e.addConfig {
					fileCreateOrAppend("werf.yaml", minimalConfigContent)
				}

				if e.commitConfig {
					gitAddAndCommit("werf.yaml")
				}

				if e.changeConfigAfterCommit {
					fileCreateOrAppend("werf.yaml", "\n")
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
			Entry("the config file not found", entry{
				expectedErrSubstring: `unable to read werf config: the file "werf.yaml" not found in the project git repository`,
			}),
			Entry("the config file not committed", entry{
				addConfig:            true,
				expectedErrSubstring: `unable to read werf config: the file "werf.yaml" must be committed`,
			}),
			Entry("the config file committed", entry{
				addConfig:    true,
				commitConfig: true,
			}),
			Entry("the config file changed after commit", entry{
				addConfig:               true,
				commitConfig:            true,
				changeConfigAfterCommit: true,
				expectedErrSubstring:    `unable to read werf config: the file "werf.yaml" changes must be committed`,
			}),
			Entry("config.allowUncommitted allows not committed config file", entry{
				allowUncommitted: true,
				addConfig:        true,
			}),
			Entry("config.allowUncommitted allows committed file", entry{
				allowUncommitted: true,
				addConfig:        true,
				commitConfig:     true,
			}),
		)
	})
})
