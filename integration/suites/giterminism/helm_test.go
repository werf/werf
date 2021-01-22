package giterminism_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config", func() {
	BeforeEach(func() {
		gitInit()
		utils.CopyIn(utils.FixturePath("default"), SuiteData.TestDirPath)
		gitAddAndCommit("werf.yaml")
		gitAddAndCommit("werf-giterminism.yaml")
	})

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
			expectedErrSubstring: `the chart directory ".helm" not found in the project git repository`,
		}),
		Entry(`the template file ".helm/templates/template1.yaml" not committed`, entry{
			addFiles:             []string{".helm/templates/template1.yaml"},
			expectedErrSubstring: `the uncommitted configuration found in the project git work tree: the chart file ".helm/templates/template1.yaml" must be committed`,
		}),
		Entry("the template files not committed", entry{
			addFiles:    []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
			commitFiles: []string{".helm/templates/template1.yaml"},
			expectedErrSubstring: `the uncommitted configuration found in the project git work tree: the following chart files must be committed:

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
			expectedErrSubstring:   `the uncommitted configuration found in the project git work tree: the chart file ".helm/templates/template1.yaml" changes must be committed`,
		}),
		Entry("the template files changed after commit", entry{
			addFiles:               []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
			commitFiles:            []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
			changeFilesAfterCommit: []string{".helm/templates/template1.yaml", ".helm/templates/template2.yaml", ".helm/templates/template3.yaml"},
			expectedErrSubstring: `the uncommitted configuration found in the project git work tree: the following chart files changes must be committed:

 - .helm/templates/template1.yaml
 - .helm/templates/template2.yaml
 - .helm/templates/template3.yaml

`,
		}),
		Entry("helm.allowUncommittedFiles (.helm/**/*) covers the not committed template", entry{
			allowUncommittedFilesGlob: ".helm/**/*",
			addFiles:                  []string{".helm/templates/template1.yaml"},
		}),
	)
})
