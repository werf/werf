package giterminism_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config go template", func() {
	BeforeEach(CommonBeforeEach)

	{
		type entryBase struct {
			allowUncommittedFilesGlob string
			addFiles                  []string
			commitFiles               []string
			changeFilesAfterCommit    []string
			expectedErrSubstring      string
		}

		bodyFuncBase := func(e entryBase) {
			if e.allowUncommittedFilesGlob != "" {
				contentToAppend := fmt.Sprintf(`
config:
  goTemplateRendering:
    allowUncommittedFiles: [%s]`, e.allowUncommittedFilesGlob)
				fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
				gitAddAndCommit("werf-giterminism.yaml")
			}

			for _, relPath := range e.addFiles {
				fileCreateOrAppend(relPath, fmt.Sprintf("# %s", relPath))
			}

			for _, relPath := range e.commitFiles {
				gitAddAndCommit(relPath)
			}

			for _, relPath := range e.changeFilesAfterCommit {
				fileCreateOrAppend(relPath, "\n")
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

				for _, relPath := range e.addFiles {
					Ω(string(output)).Should(ContainSubstring(fmt.Sprintf("# %s", relPath)))
				}
			}
		}

		Context(".Files.Get", func() {
			DescribeTable("config.goTemplateRendering.allowUncommittedFiles",
				func(e entryBase) {
					fileCreateOrAppend("werf.yaml", fmt.Sprintf(`{{ .Files.Get "%s" }}`, ".werf/file"))
					gitAddAndCommit("werf.yaml")

					bodyFuncBase(e)
				},
				Entry("the file not found", entryBase{
					expectedErrSubstring: "error calling Get: {{ .Files.Get '.werf/file' }}: the file '.werf/file' not found in the project git repository",
				}),
				Entry("the file not committed", entryBase{
					addFiles:             []string{".werf/file"},
					expectedErrSubstring: "error calling Get: {{ .Files.Get '.werf/file' }}: the uncommitted configuration found in the project directory: the file '.werf/file' must be committed",
				}),
				Entry("the file committed", entryBase{
					addFiles:    []string{".werf/file"},
					commitFiles: []string{".werf/file"},
				}),
				Entry("config.goTemplateRendering.allowUncommittedFiles covers the not committed file", entryBase{
					allowUncommittedFilesGlob: ".werf/file",
					addFiles:                  []string{".werf/file"},
				}),
				Entry("config.goTemplateRendering.allowUncommittedFiles covers the committed file", entryBase{
					allowUncommittedFilesGlob: ".werf/file",
					addFiles:                  []string{".werf/file"},
					commitFiles:               []string{".werf/file"},
				}),
				Entry("the file committed, the file has uncommitted changes", entryBase{
					addFiles:               []string{".werf/file"},
					commitFiles:            []string{".werf/file"},
					changeFilesAfterCommit: []string{".werf/file"},
					expectedErrSubstring:   "error calling Get: {{ .Files.Get '.werf/file' }}: the uncommitted configuration found in the project directory: the file '.werf/file' changes must be committed",
				}),
			)
		})

		Context(".Files.Glob", func() {
			type entry struct {
				filesGlob string
				entryBase
			}

			DescribeTable("config.goTemplateRendering.allowUncommittedFiles",
				func(e entry) {
					fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
{{ range $path, $content := .Files.Glob "%s" }}
{{ $content }}
{{ end }}
`, e.filesGlob))
					gitAddAndCommit("werf.yaml")

					bodyFuncBase(e.entryBase)
				},
				Entry("nothing found", entry{}),
				Entry("the file1 not committed", entry{
					filesGlob: ".werf/file1",
					entryBase: entryBase{
						addFiles:             []string{".werf/file1"},
						expectedErrSubstring: "error calling Glob: {{ .Files.Glob '.werf/file1' }}: the uncommitted configuration found in the project directory: the file '.werf/file1' must be committed",
					},
				}),
				Entry("the files not committed", entry{
					filesGlob: ".werf/*",
					entryBase: entryBase{
						addFiles: []string{".werf/file1", ".werf/file2", ".werf/file3"},
						expectedErrSubstring: `error calling Glob: {{ .Files.Glob '.werf/*' }}: the uncommitted configuration found in the project directory: the following files must be committed:

 - .werf/file1
 - .werf/file2
 - .werf/file3

`,
					},
				}),
				Entry("the file1 committed", entry{
					filesGlob: ".werf/file1",
					entryBase: entryBase{
						addFiles:    []string{".werf/file1"},
						commitFiles: []string{".werf/file1"},
					},
				}),
				Entry("the file1 committed, the file1 has uncommitted changes", entry{
					filesGlob: ".werf/file1",
					entryBase: entryBase{
						addFiles:               []string{".werf/file1"},
						commitFiles:            []string{".werf/file1"},
						changeFilesAfterCommit: []string{".werf/file1"},
						expectedErrSubstring:   "error calling Glob: {{ .Files.Glob '.werf/file1' }}: the uncommitted configuration found in the project directory: the file '.werf/file1' changes must be committed",
					},
				}),
				Entry("config.goTemplateRendering.allowUncommittedFiles (.werf/file1) covers the not committed file", entry{
					filesGlob: ".werf/file1",
					entryBase: entryBase{
						allowUncommittedFilesGlob: ".werf/file1",
						addFiles:                  []string{".werf/file1"},
					},
				}),
				Entry("config.goTemplateRendering.allowUncommittedFiles (.werf/file1) covers the committed file", entry{
					filesGlob: ".werf/file1",
					entryBase: entryBase{
						allowUncommittedFilesGlob: ".werf/file1",
						addFiles:                  []string{".werf/file1"},
						changeFilesAfterCommit:    []string{".werf/file1"},
					},
				}),
				Entry("config.goTemplateRendering.allowUncommittedFiles (/**/*/) covers the not committed files", entry{
					filesGlob: ".werf/*",
					entryBase: entryBase{
						allowUncommittedFilesGlob: `/**/*/`,
						addFiles:                  []string{".werf/file1", ".werf/file2", ".werf/file3"},
					},
				}),
			)
		})
	}
})
