package giterminism_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("werf.yaml", func() {
	BeforeEach(ConfigBeforeEach)

	type entry struct {
		allowUncommitted        bool
		commitConfig            bool
		changeConfigAfterCommit bool
		expectedErrSubstring    string
	}

	DescribeTable("allowUncommitted",
		func(e entry) {
			configPath := filepath.Join(SuiteData.TestDirPath, "werf.yaml")
			giterminismConfigPath := filepath.Join(SuiteData.TestDirPath, "werf-giterminism.yaml")

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
			fileCreateOrAppend(giterminismConfigPath, contentToAppend)
			gitAddAndCommit("werf-giterminism.yaml")

			if e.commitConfig {
				gitAddAndCommit("werf.yaml")
			}

			if e.changeConfigAfterCommit {
				fileCreateOrAppend(configPath, "\n")
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
			allowUncommitted:     false,
			commitConfig:         false,
			expectedErrSubstring: "the werf config 'werf.yaml', 'werf.yml' not found in the project git repository",
		}),
		Entry("werf.yaml committed", entry{
			allowUncommitted: false,
			commitConfig:     true,
		}),
		Entry("config.allowUncommitted is true, werf.yaml not committed", entry{
			allowUncommitted: true,
			commitConfig:     false,
		}),
		Entry("config.allowUncommitted is true, werf.yaml committed", entry{
			allowUncommitted: true,
			commitConfig:     true,
		}),
		Entry("werf.yaml committed, local werf.yaml has uncommitted changes", entry{
			allowUncommitted:        false,
			commitConfig:            true,
			changeConfigAfterCommit: true,
			expectedErrSubstring:    "the werf config 'werf.yaml' must be committed",
		}),
	)
})
