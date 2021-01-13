package giterminism_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe(".Files.Get", func() {
	BeforeEach(CommonBeforeEach)

	Context("config.goTemplateRendering.allowUncommittedFiles", func() {
		type entry struct {
			allowUncommittedFile  bool
			addFile               bool
			commitFile            bool
			changeFileAfterCommit bool
			expectedErrSubstring  string
		}

		DescribeTable(".Files.Get",
			func(e entry) {
				relFilePath := "test"
				filePath := filepath.Join(SuiteData.TestDirPath, relFilePath)
				configPath := filepath.Join(SuiteData.TestDirPath, "werf.yaml")
				giterminismConfigPath := filepath.Join(SuiteData.TestDirPath, "werf-giterminism.yaml")

				fileCreateOrAppend(configPath, `{{ .Files.Get "test" }}`)
				gitAddAndCommit("werf.yaml")

				if e.allowUncommittedFile {
					contentToAppend := `
config:
  goTemplateRendering:
    allowUncommittedFiles: [test]`
					fileCreateOrAppend(giterminismConfigPath, contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
				}

				if e.addFile {
					fileCreateOrAppend(filePath, `
# test
`)
				}

				if e.commitFile {
					gitAddAndCommit(relFilePath)
				}

				if e.changeFileAfterCommit {
					fileCreateOrAppend(filePath, "\n")
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

					if e.addFile {
						Ω(string(output)).Should(ContainSubstring("# test"))
					}
				}
			},
			Entry("the file not found", entry{
				expectedErrSubstring: "the file 'test' not found in the project git repository",
			}),
			Entry("the file not committed", entry{
				addFile:              true,
				expectedErrSubstring: "the file 'test' not found in the project git repository",
			}),
			Entry("the file committed", entry{
				addFile:    true,
				commitFile: true,
			}),
			Entry("config.goTemplateRendering.allowUncommittedFiles has the not committed file", entry{
				allowUncommittedFile: true,
				addFile:              true,
			}),
			Entry("config.goTemplateRendering.allowUncommittedFiles has the committed file", entry{
				allowUncommittedFile: true,
				addFile:              true,
				commitFile:           true,
			}),
			Entry("the file committed, the file has uncommitted changes", entry{
				addFile:               true,
				commitFile:            true,
				changeFileAfterCommit: true,
				expectedErrSubstring:  "the file 'test' must be committed",
			}),
		)
	})
})
