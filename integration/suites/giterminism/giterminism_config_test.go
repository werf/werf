package giterminism_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("giterminism config", func() {
	BeforeEach(func() {
		gitInit()
		utils.CopyIn(utils.FixturePath("giterminism_config"), SuiteData.TestDirPath)
		gitAddAndCommit("werf.yaml")
	})

	type entry struct {
		addConfig               bool
		commitConfig            bool
		changeConfigAfterCommit bool
		expectedErrSubstring    string
	}

	DescribeTable("",
		func(e entry) {
			if e.addConfig {
				contentToAppend := `giterminismConfigVersion: "1"`
				fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
			}

			if e.commitConfig {
				gitAddAndCommit("werf-giterminism.yaml")
			}

			if e.changeConfigAfterCommit {
				fileCreateOrAppend("werf-giterminism.yaml", "\n")
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
		Entry("the giterminism config not exist", entry{}),
		Entry("the giterminism config not committed", entry{
			addConfig:            true,
			expectedErrSubstring: "the uncommitted configuration found in the project directory: the giterminism config 'werf-giterminism.yaml' must be committed",
		}),
		Entry("the giterminism config committed", entry{
			addConfig:    true,
			commitConfig: true,
		}),
		Entry("the giterminism config changed after commit", entry{
			addConfig:               true,
			commitConfig:            true,
			changeConfigAfterCommit: true,
			expectedErrSubstring:    "the uncommitted configuration found in the project directory: the giterminism config 'werf-giterminism.yaml' changes must be committed",
		}),
	)
})
