package giterminism_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("config templates dir", func() {
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

	DescribeTable("config.allowUncommittedTemplates",
		func(ctx SpecContext, e entry) {
			tmpl1RelPath := ".werf/templates/1.tmpl"
			tmpl2RelPath := ".werf/templates/2.tmpl"

			var contentToAppend string
			if e.allowUncommittedTemplate1 {
				contentToAppend = `
config:
  allowUncommittedTemplates: [.werf/templates/1.tmpl]`
			} else if e.allowUncommittedAllTemplates {
				contentToAppend = `
config:
  allowUncommittedTemplates: [".werf/**/*.tmpl"]`
			}

			if contentToAppend != "" {
				fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
				gitAddAndCommit(ctx, "werf-giterminism.yaml")
			}

			if e.addTemplate1 {
				fileCreateOrAppend(tmpl1RelPath, `
# template .werf/templates/1.tmpl
`)
				fileCreateOrAppend("werf.yaml", `{{ include "templates/1.tmpl" . }}`)
				gitAddAndCommit(ctx, "werf.yaml")
			}

			if e.commitTemplate1 {
				gitAddAndCommit(ctx, tmpl1RelPath)
			}

			if e.addTemplate2 {
				fileCreateOrAppend(tmpl2RelPath, `
# template .werf/templates/2.tmpl
`)
				fileCreateOrAppend("werf.yaml", `{{ include "templates/2.tmpl" . }}`)
				gitAddAndCommit(ctx, "werf.yaml")
			}

			if e.changeTemplate1AfterCommit {
				fileCreateOrAppend(tmpl1RelPath, "\n")
			}

			output, err := utils.RunCommand(
				ctx,
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"config", "render",
			)

			if e.expectedErrSubstring != "" {
				Expect(err).Should(HaveOccurred())
				Expect(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
			} else {
				Expect(err).ShouldNot(HaveOccurred())

				if e.addTemplate1 {
					Expect(string(output)).Should(ContainSubstring("# template .werf/templates/1.tmpl"))
				}

				if e.addTemplate2 {
					Expect(string(output)).Should(ContainSubstring("# template .werf/templates/2.tmpl"))
				}
			}
		},
		Entry(".werf/templates/1.tmpl not tracked", entry{
			addTemplate1:         true,
			expectedErrSubstring: `unable to read werf config templates: the untracked file ".werf/templates/1.tmpl" must be committed`,
		}),
		Entry(".werf/templates/1.tmpl committed", entry{
			addTemplate1:    true,
			commitTemplate1: true,
		}),
		Entry(".werf/templates/1.tmpl committed, the template file has uncommitted changes", entry{
			addTemplate1:               true,
			commitTemplate1:            true,
			changeTemplate1AfterCommit: true,
			expectedErrSubstring:       `unable to read werf config templates: the file ".werf/templates/1.tmpl" must be committed`,
		}),
		Entry("config.allowUncommittedTemplates has .werf/templates/1.tmpl, the template file not tracked", entry{
			allowUncommittedTemplate1: true,
			addTemplate1:              true,
		}),
		Entry("config.allowUncommittedTemplates has .werf/templates/1.tmpl, the template file committed", entry{
			allowUncommittedTemplate1: true,
			addTemplate1:              true,
			commitTemplate1:           true,
		}),
		Entry("config.allowUncommittedTemplates has .werf/templates/1.tmpl, .werf/templates/2.tmpl not tracked", entry{
			allowUncommittedTemplate1: true,
			addTemplate1:              true,
			addTemplate2:              true,
			commitTemplate1:           true,
			expectedErrSubstring:      `unable to read werf config templates: the untracked file ".werf/templates/2.tmpl" must be committed`,
		}),
		Entry("config.allowUncommittedTemplates has .werf/**/*.tmpl", entry{
			allowUncommittedAllTemplates: true,
			addTemplate1:                 true,
			addTemplate2:                 true,
		}),
	)
})
