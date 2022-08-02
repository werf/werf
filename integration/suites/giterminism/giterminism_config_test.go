package giterminism_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
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
		Entry("the giterminism config not tracked", entry{
			addConfig:            true,
			expectedErrSubstring: `unable to read werf giterminism config: the untracked file "werf-giterminism.yaml" must be committed`,
		}),
		Entry("the giterminism config committed", entry{
			addConfig:    true,
			commitConfig: true,
		}),
		Entry("the giterminism config changed after commit", entry{
			addConfig:               true,
			commitConfig:            true,
			changeConfigAfterCommit: true,
			expectedErrSubstring:    `unable to read werf giterminism config: the file "werf-giterminism.yaml" must be committed`,
		}),
	)
})
