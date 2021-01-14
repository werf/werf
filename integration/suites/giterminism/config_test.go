package giterminism_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config", func() {
	BeforeEach(ConfigBeforeEach)

	type entry struct {
		allowUncommitted        bool
		commitConfig            bool
		changeConfigAfterCommit bool
		expectedErrSubstring    string
	}

	DescribeTable("allowUncommitted",
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
		Entry("werf.yaml not found in commit", entry{
			expectedErrSubstring: `the following werf configs not found in the project git repository:

 - werf.yaml
 - werf.yml

`,
		}),
		Entry("werf.yaml committed", entry{
			commitConfig: true,
		}),
		Entry("werf.yaml committed, werf.yaml has uncommitted changes", entry{
			commitConfig:            true,
			changeConfigAfterCommit: true,
			expectedErrSubstring:    `the uncommitted configuration found in the project directory: the werf config 'werf.yaml' changes must be committed`,
		}),
		Entry("config.allowUncommitted is true, werf.yaml not committed", entry{
			allowUncommitted: true,
		}),
		Entry("config.allowUncommitted is true, werf.yaml committed", entry{
			allowUncommitted: true,
			commitConfig:     true,
		}),
	)
})
