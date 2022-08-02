package giterminism_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("helm chart files", func() {
	for _, value := range []string{"", "dir/custom_project_dir"} {
		projectRelPath := value

		relativeToProjectDir := func(path string) string {
			return filepath.Join(projectRelPath, path)
		}

		werfGiterminismRelPath := relativeToProjectDir("werf-giterminism.yaml")
		werfConfigRelPath := relativeToProjectDir("werf.yaml")

		type symlinkEntry struct {
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

		symlinkBody := func(e symlinkEntry) {
			if e.skipOnWindows && runtime.GOOS == "windows" {
				Skip("skip on windows")
			}

			{ // werf-giterminism.yaml
				if len(e.allowUncommittedFilesGlobs) != 0 {
					contentToAppend := fmt.Sprintf(`
helm:
  allowUncommittedFiles: ["%s"]
`, strings.Join(e.allowUncommittedFilesGlobs, `", "`))
					fileCreateOrAppend(werfGiterminismRelPath, contentToAppend)
					gitAddAndCommit(werfGiterminismRelPath)
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
				filepath.Join(SuiteData.TestDirPath, projectRelPath),
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
		}

		runCommonTests := func() {
			BeforeEach(func() {
				gitInit()
				utils.CopyIn(utils.FixturePath("default"), filepath.Join(SuiteData.TestDirPath, projectRelPath))
				gitAddAndCommit(werfGiterminismRelPath)
				gitAddAndCommit(werfConfigRelPath)
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
							fileCreateOrAppend(werfGiterminismRelPath, contentToAppend)
							gitAddAndCommit(werfGiterminismRelPath)
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
							filepath.Join(SuiteData.TestDirPath, projectRelPath),
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
					Entry(`the template file ".helm/templates/template1.yaml" not tracked`, entry{
						addFiles:             []string{relativeToProjectDir(".helm/templates/template1.yaml")},
						expectedErrSubstring: `unable to locate chart directory: the untracked file ".helm/templates/template1.yaml" must be committed`,
					}),
					Entry("the template files not tracked", entry{
						addFiles: []string{
							relativeToProjectDir(".helm/templates/template1.yaml"),
							relativeToProjectDir(".helm/templates/template2.yaml"),
							relativeToProjectDir(".helm/templates/template3.yaml"),
						},
						commitFiles: []string{relativeToProjectDir(".helm/templates/template1.yaml")},
						expectedErrSubstring: `unable to locate chart directory: the following untracked files must be committed:

 - .helm/templates/template2.yaml
 - .helm/templates/template3.yaml

`,
					}),
					Entry(`the template file ".helm/templates/template1.yaml" committed`, entry{
						addFiles:    []string{relativeToProjectDir(".helm/templates/template1.yaml")},
						commitFiles: []string{relativeToProjectDir(".helm/templates/template1.yaml")},
					}),
					Entry(`the template file ".helm/templates/template1.yaml" changed after commit`, entry{
						addFiles:               []string{relativeToProjectDir(".helm/templates/template1.yaml")},
						commitFiles:            []string{relativeToProjectDir(".helm/templates/template1.yaml")},
						changeFilesAfterCommit: []string{relativeToProjectDir(".helm/templates/template1.yaml")},
						expectedErrSubstring:   `unable to locate chart directory: the file ".helm/templates/template1.yaml" must be committed`,
					}),
					Entry("the template files changed after commit", entry{
						addFiles: []string{
							relativeToProjectDir(".helm/templates/template1.yaml"),
							relativeToProjectDir(".helm/templates/template2.yaml"),
							relativeToProjectDir(".helm/templates/template3.yaml"),
						},
						commitFiles: []string{
							relativeToProjectDir(".helm/templates/template1.yaml"),
							relativeToProjectDir(".helm/templates/template2.yaml"),
							relativeToProjectDir(".helm/templates/template3.yaml"),
						},
						changeFilesAfterCommit: []string{
							relativeToProjectDir(".helm/templates/template1.yaml"),
							relativeToProjectDir(".helm/templates/template2.yaml"),
							relativeToProjectDir(".helm/templates/template3.yaml"),
						},
						expectedErrSubstring: `unable to locate chart directory: the following files must be committed:

 - .helm/templates/template1.yaml
 - .helm/templates/template2.yaml
 - .helm/templates/template3.yaml

`,
					}),
					Entry("helm.allowUncommittedFiles (.helm/**/*) covers the not tracked template", entry{
						allowUncommittedFilesGlob: ".helm/**/*",
						addFiles:                  []string{relativeToProjectDir(".helm/templates/template1.yaml")},
					}))
			})

			Context("symlinks", func() {
				DescribeTable("helm.allowUncommittedFiles",
					symlinkBody,
					Entry("the symlink to the chart dir: .helm -> dir/.helm", symlinkEntry{
						skipOnWindows: true,
						addFiles:      []string{relativeToProjectDir("dir/.helm/templates/template1.yaml")},
						commitFiles:   []string{relativeToProjectDir("dir/.helm/templates/template1.yaml")},
						addAndCommitSymlinks: map[string]string{
							relativeToProjectDir(".helm"): getLinkTo(relativeToProjectDir(".helm"), relativeToProjectDir("dir/.helm")),
						},
					}),
					Entry("the symlink to the chart templates dir: .helm/templates -> dir/.helm/templates", symlinkEntry{
						skipOnWindows: true,
						addFiles:      []string{relativeToProjectDir("dir/.helm/templates/template1.yaml")},
						commitFiles:   []string{relativeToProjectDir("dir/.helm/templates/template1.yaml")},
						addAndCommitSymlinks: map[string]string{
							relativeToProjectDir(".helm/templates"): getLinkTo(relativeToProjectDir(".helm/templates"), relativeToProjectDir("dir/.helm/templates")),
						},
					}),
					Entry("helm.allowUncommittedFiles (.helm) does not cover uncommitted files", symlinkEntry{
						skipOnWindows:              true,
						allowUncommittedFilesGlobs: []string{".helm"},
						addFiles:                   []string{relativeToProjectDir("dir/.helm/templates/template1.yaml")},
						addSymlinks: map[string]string{
							relativeToProjectDir(".helm/templates"): getLinkTo(relativeToProjectDir(".helm/templates"), relativeToProjectDir("dir/.helm/templates")),
						},
						expectedErrSubstring: `unable to locate chart directory: accepted file ".helm/templates/template1.yaml" check failed: the link target "dir/.helm/templates" should be also accepted by giterminism config`,
					}),
					Entry("helm.allowUncommittedFiles (.helm, dir) covers uncommitted files", symlinkEntry{
						skipOnWindows:              true,
						allowUncommittedFilesGlobs: []string{".helm", "dir"},
						addFiles:                   []string{relativeToProjectDir("dir/.helm/templates/template1.yaml")},
						addSymlinks: map[string]string{
							relativeToProjectDir(".helm/templates"): getLinkTo(relativeToProjectDir(".helm/templates"), relativeToProjectDir("dir/.helm/templates")),
						},
					}),
					Entry("the allowed symlink to the chart template and the committed one", symlinkEntry{
						skipOnWindows: true,
						allowUncommittedFilesGlobs: []string{
							".helm/templates/template1.yaml",
							"dir/.helm/templates/template1.yaml",
						},
						addFiles: []string{
							relativeToProjectDir("dir/.helm/templates/template1.yaml"),
							relativeToProjectDir(".helm/templates/template2.yaml"),
						},
						commitFiles: []string{relativeToProjectDir(".helm/templates/template2.yaml")},
						addSymlinks: map[string]string{
							relativeToProjectDir(".helm/templates/template1.yaml"): getLinkTo(relativeToProjectDir(".helm/templates/template1.yaml"), relativeToProjectDir("dir/.helm/templates/template1.yaml")),
						},
					}))
			})
		}

		if projectRelPath == "" {
			{
				runCommonTests()
			}
		} else {
			Context("custom project directory", func() {
				runCommonTests()

				It("the symlink to the chart dir outside project directory: .helm -> ../../helm", func() {
					symlinkBody(symlinkEntry{
						skipOnWindows: true, // TODO does not appear to be a gzipped archive; got 'application/octet-stream'
						addFiles:      []string{"helm/templates/template1.yaml"},
						commitFiles:   []string{"helm/templates/template1.yaml"},
						addAndCommitSymlinks: map[string]string{
							relativeToProjectDir(".helm"): getLinkTo(relativeToProjectDir(".helm"), "helm"),
						},
					})
				})

				It("the symlink to the uncommitted directory outside project directory: .helm -> ../../helm", func() {
					symlinkBody(symlinkEntry{
						skipOnWindows: true,
						addFiles:      []string{"helm/templates/template1.yaml"},
						addAndCommitSymlinks: map[string]string{
							relativeToProjectDir(".helm"): getLinkTo(relativeToProjectDir(".helm"), "helm"),
						},
						expectedErrSubstring: `unable to locate chart directory: the file ".helm/templates/template1.yaml" not found in the project git repository`,
					})
				})

				It("helm.allowUncommittedFiles (.helm) does not cover uncommitted files: .helm -> ../../helm", func() {
					symlinkBody(symlinkEntry{
						skipOnWindows:              true,
						allowUncommittedFilesGlobs: []string{".helm"},
						addFiles:                   []string{"helm/templates/template1.yaml"},
						addSymlinks: map[string]string{
							relativeToProjectDir(".helm"): getLinkTo(relativeToProjectDir(".helm"), "helm"),
						},
						expectedErrSubstring: `unable to locate chart directory: accepted file ".helm/templates/template1.yaml" check failed: the link target "../../helm" should be also accepted by giterminism config`,
					})
				})

				It("helm.allowUncommittedFiles (.helm, ../../helm) covers uncommitted files: .helm -> ../../helm", func() {
					symlinkBody(symlinkEntry{
						skipOnWindows:              true,
						allowUncommittedFilesGlobs: []string{".helm", "../../helm"},
						addFiles:                   []string{"helm/templates/template1.yaml"},
						addSymlinks: map[string]string{
							relativeToProjectDir(".helm"): getLinkTo(relativeToProjectDir(".helm"), "helm"),
						},
					})
				})

				It("helm.allowUncommittedFiles (**/*) covers uncommitted files: .helm -> ../../helm", func() {
					symlinkBody(symlinkEntry{
						skipOnWindows:              true,
						allowUncommittedFilesGlobs: []string{"**/*"},
						addFiles:                   []string{"helm/templates/template1.yaml"},
						addSymlinks: map[string]string{
							relativeToProjectDir(".helm"): getLinkTo(relativeToProjectDir(".helm"), "helm"),
						},
					})
				})
			})
		}
	}
})
