package render_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

func werfRender(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"render"}, extraArgs...)...)...)
}

var _ = Describe("Helm values yaml parser", func() {
	Context("when values.yaml contains anchors and duplicated fields, https://github.com/werf/werf/issues/2871", func() {
		It("should parse values.yaml and override data by duplicate keys", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "helm_values_yaml_parser", "initial commit")

			gotTestLine := false
			Expect(werfRender(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "test1: override") {
						gotTestLine = true
					}
				},
			})).Should(Succeed())

			Expect(gotTestLine).Should(BeTrue())
		})
	})
})
