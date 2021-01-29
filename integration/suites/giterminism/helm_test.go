package giterminism_test

import (
	"fmt"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("helm chart files", func() {
	BeforeEach(func() {
		gitInit()
		utils.CopyIn(utils.FixturePath("default"), SuiteData.TestDirPath)
		gitAddAndCommit("werf.yaml")
		gitAddAndCommit("werf-giterminism.yaml")
	})

	Context("regular files", func() {
		type entry struct {
			allowUncommittedFilesGlob string
			addFiles                  []string
			commitFiles               []string
			changeFilesAfterCommit    []string
			expectedErrSubstring      string
		}

		DescribeTable("helm.allowUncommittedFiles",
			func(e entry) {
				var contentToAppend string
				if e.allowUncommittedFilesGlob != "" {
					contentToAppend = fmt.Sprintf(`
helm:
  allowUncommittedFiles: ["%s"]`, e.allowUncommittedFilesGlob)
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
				}

				for _, relPath := range e.addFiles {
					fileCreateOrAppend(relPath, fmt.Sprintf(`test: %s`, relPath))
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
					"render",
				)

				if e.expectedErrSubstring != "" {
					Ω(err).Should(HaveOccurred())
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(err).ShouldNot(HaveOccurred())

					for _, relPath := range e.addFiles {
						Ω(string(output)).Should(ContainSubstring(fmt.Sprintf(`test: %s`, relPath)))
					}
				}
			},
			Entry("the chart directory not found", entry{
				expectedErrSubstring: `unable to locate chart directory: the directory ".helm" not found in the project git repository`,
			}),
			Entry(`the template file ".helm/templates/template1.yaml" not committed`, entry{
				addFiles:             []string{".helm/templates/template1.yaml"},
				expectedErrSubstring: `unable to locate chart directory: the file ".helm/templates/template1.yaml" must be committed`,
			}),
			Entry("the template files not committed", entry{
				addFiles:    []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
				commitFiles: []string{".helm/templates/template1.yaml"},
				expectedErrSubstring: `unable to locate chart directory: the following files must be committed:

 - .helm/templates/template2.yaml
 - .helm/templates/template3.yaml

`,
			}),
			Entry(`the template file ".helm/templates/template1.yaml" committed`, entry{
				addFiles:    []string{".helm/templates/template1.yaml"},
				commitFiles: []string{".helm/templates/template1.yaml"},
			}),
			Entry(`the template file ".helm/templates/template1.yaml" changed after commit`, entry{
				addFiles:               []string{".helm/templates/template1.yaml"},
				commitFiles:            []string{".helm/templates/template1.yaml"},
				changeFilesAfterCommit: []string{".helm/templates/template1.yaml"},
				expectedErrSubstring:   `unable to locate chart directory: the file ".helm/templates/template1.yaml" must be committed`,
			}),
			Entry("the template files changed after commit", entry{
				addFiles:               []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
				commitFiles:            []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
				changeFilesAfterCommit: []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
				expectedErrSubstring: `unable to locate chart directory: the following files must be committed:

 - .helm/templates/template1.yaml
 - .helm/templates/template2.yaml
 - .helm/templates/template3.yaml

`,
			}),
			Entry("helm.allowUncommittedFiles (.helm/**/*) covers the not committed template", entry{
				allowUncommittedFilesGlob: ".helm/**/*",
				addFiles:                  []string{".helm/templates/template1.yaml"},
			}))
	})

	Context("symlinks", func() {
		type entry struct {
			allowUncommittedFilesGlobs []string
			addFiles                   []string
			commitFiles                []string
			changeFilesAfterCommit     []string
			addSymlinks                map[string]string
			addAndCommitSymlinks       map[string]string
			changeSymlinksAfterCommit  map[string]string
			expectedErrSubstring       string
			skipOnWindows              bool
		}

		DescribeTable("helm.allowUncommittedFiles",
			func(e entry) {
				if e.skipOnWindows && runtime.GOOS == "windows" {
					Skip("skip on windows")
				}

				{ // werf-giterminism.yaml
					if len(e.allowUncommittedFilesGlobs) != 0 {
						contentToAppend := fmt.Sprintf(`
helm:
  allowUncommittedFiles: ["%s"]
`, strings.Join(e.allowUncommittedFilesGlobs, `", "`))
						fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
						gitAddAndCommit("werf-giterminism.yaml")
					}
				}

				{ // helm files
					for _, relPath := range e.addFiles {
						fileCreateOrAppend(relPath, fmt.Sprintf(`test: %s`, relPath))
					}

					for _, relPath := range e.commitFiles {
						gitAddAndCommit(relPath)
					}

					for _, relPath := range e.changeFilesAfterCommit {
						fileCreateOrAppend(relPath, "\n")
					}

					for path, link := range e.addSymlinks {
						symlinkFileCreateOrModify(path, link)
					}

					for path, link := range e.addAndCommitSymlinks {
						symlinkFileCreateOrModifyAndAdd(path, link)
						gitAddAndCommit(path)
					}

					for path, link := range e.changeSymlinksAfterCommit {
						symlinkFileCreateOrModify(path, link)
					}
				}

				output, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"render",
				)

				if e.expectedErrSubstring != "" {
					Ω(err).Should(HaveOccurred())
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(err).ShouldNot(HaveOccurred())

					for _, relPath := range e.addFiles {
						Ω(string(output)).Should(ContainSubstring(fmt.Sprintf(`test: %s`, relPath)))
					}
				}
			},
			Entry("the symlink to the chart dir: .helm -> dir/.helm", entry{
				skipOnWindows: true,
				addFiles:      []string{"dir/.helm/templates/template1.yaml"},
				commitFiles:   []string{"dir/.helm/templates/template1.yaml"},
				addAndCommitSymlinks: map[string]string{
					".helm": "dir/.helm",
				},
			}),
			Entry("the symlink to the chart templates dir: .helm/templates -> dir/.helm/templates", entry{
				skipOnWindows: true,
				addFiles:      []string{"dir/.helm/templates/template1.yaml"},
				commitFiles:   []string{"dir/.helm/templates/template1.yaml"},
				addAndCommitSymlinks: map[string]string{
					".helm/templates": "../dir/.helm/templates",
				},
			}),
			Entry("helm.allowUncommittedFiles (.helm) does not cover uncommitted files", entry{
				skipOnWindows:              true,
				allowUncommittedFilesGlobs: []string{".helm"},
				addFiles:                   []string{"dir/.helm/templates/template1.yaml"},
				addSymlinks: map[string]string{
					".helm/templates": "../dir/.helm/templates",
				},
				expectedErrSubstring: `unable to locate chart directory: accepted symlink ".helm/templates/template1.yaml" check failed: the link target "dir/.helm/templates" should be also accepted by giterminism config`,
			}),
			Entry("helm.allowUncommittedFiles (.helm, dir) covers uncommitted files", entry{
				skipOnWindows:              true,
				allowUncommittedFilesGlobs: []string{".helm", "dir"},
				addFiles:                   []string{"dir/.helm/templates/template1.yaml"},
				addSymlinks: map[string]string{
					".helm/templates": "../dir/.helm/templates",
				},
			}),
			Entry("the allowed symlink to the chart template and the committed one", entry{
				skipOnWindows:              true,
				allowUncommittedFilesGlobs: []string{".helm/templates/template1.yaml", "dir/.helm/templates/template1.yaml"},
				addFiles:                   []string{"dir/.helm/templates/template1.yaml", ".helm/templates/template2.yaml"},
				commitFiles:                []string{".helm/templates/template2.yaml"},
				addSymlinks: map[string]string{
					".helm/templates/template1.yaml": "../../dir/.helm/templates/template1.yaml",
				},
			}))
	})
})
