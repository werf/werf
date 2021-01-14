package giterminism_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe(".werf/**/*.tmpl", func() {
	BeforeEach(CommonBeforeEach)

	type entry struct {
		allowUncommittedTemplate1    bool
		allowUncommittedAllTemplates bool
		addTemplate1                 bool
		addTemplate2                 bool
		commitTemplate1              bool
		changeTemplate1AfterCommit   bool
		expectedErrSubstring         string
	}

	DescribeTable("allowUncommittedTemplates",
		func(e entry) {
			tmpl1RelPath := ".werf/templates/1.tmpl"
			tmpl2RelPath := ".werf/templates/2.tmpl"
			tmpl1Path := filepath.Join(SuiteData.TestDirPath, tmpl1RelPath)
			tmpl2Path := filepath.Join(SuiteData.TestDirPath, tmpl2RelPath)
			configPath := filepath.Join(SuiteData.TestDirPath, "werf.yaml")
			giterminismConfigPath := filepath.Join(SuiteData.TestDirPath, "werf-giterminism.yaml")

			var contentToAppend string
			if e.allowUncommittedTemplate1 {
				contentToAppend = `
config:
  allowUncommittedTemplates: [.werf/templates/1.tmpl]`
			} else if e.allowUncommittedAllTemplates {
				contentToAppend = `
config:
  allowUncommittedTemplates: [/.werf/**/*.tmpl/]`
			}

			if contentToAppend != "" {
				fileCreateOrAppend(giterminismConfigPath, contentToAppend)
				gitAddAndCommit("werf-giterminism.yaml")
			}

			if e.addTemplate1 {
				fileCreateOrAppend(tmpl1Path, `
# template .werf/templates/1.tmpl
`)
				fileCreateOrAppend(configPath, `{{ include "templates/1.tmpl" . }}`)
				gitAddAndCommit(configPath)
			}

			if e.commitTemplate1 {
				gitAddAndCommit(tmpl1RelPath)
			}

			if e.addTemplate2 {
				fileCreateOrAppend(tmpl2Path, `
# template .werf/templates/2.tmpl
`)
				fileCreateOrAppend(configPath, `{{ include "templates/2.tmpl" . }}`)
				gitAddAndCommit(configPath)
			}

			if e.changeTemplate1AfterCommit {
				fileCreateOrAppend(tmpl1Path, "\n")
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

				if e.addTemplate1 {
					Ω(string(output)).Should(ContainSubstring("# template .werf/templates/1.tmpl"))
				}

				if e.addTemplate2 {
					Ω(string(output)).Should(ContainSubstring("# template .werf/templates/2.tmpl"))
				}
			}
		},
		Entry(".werf/templates/1.tmpl not found in commit", entry{
			addTemplate1:         true,
			expectedErrSubstring: fmt.Sprintf("the werf config template '%s' must be committed", filepath.FromSlash(".werf/templates/1.tmpl")),
		}),
		Entry(".werf/templates/1.tmpl committed", entry{
			addTemplate1:    true,
			commitTemplate1: true,
		}),
		Entry(".werf/templates/1.tmpl committed, the template file has uncommitted changes", entry{
			addTemplate1:               true,
			commitTemplate1:            true,
			changeTemplate1AfterCommit: true,
			expectedErrSubstring:       fmt.Sprintf("the werf config template '%s' must be committed", filepath.FromSlash(".werf/templates/1.tmpl")),
		}),
		Entry("config.allowUncommittedTemplates has .werf/templates/1.tmpl, the template file not committed", entry{
			allowUncommittedTemplate1: true,
			addTemplate1:              true,
		}),
		Entry("config.allowUncommittedTemplates has .werf/templates/1.tmpl, the template file committed", entry{
			allowUncommittedTemplate1: true,
			addTemplate1:              true,
			commitTemplate1:           true,
		}),
		Entry("config.allowUncommittedTemplates has .werf/templates/1.tmpl, .werf/templates/2.tmpl not committed", entry{
			allowUncommittedTemplate1: true,
			addTemplate1:              true,
			addTemplate2:              true,
			commitTemplate1:           true,
			expectedErrSubstring:      fmt.Sprintf("the werf config template '%s' must be committed", filepath.FromSlash(".werf/templates/2.tmpl")),
		}),
		Entry("config.allowUncommittedTemplates has /.werf/**/*.tmpl/", entry{
			allowUncommittedAllTemplates: true,
			addTemplate1:                 true,
			addTemplate2:                 true,
		}),
	)
})
