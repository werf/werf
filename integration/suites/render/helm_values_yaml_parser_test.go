package render_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfRender(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"render"}, extraArgs...)...)
}

var _ = Describe("Helm values yaml parser", func() {
	Context("when values.yaml contains anchors and duplicated fields, https://github.com/werf/werf/issues/2871", func() {
		It("should parse values.yaml and override data by duplicate keys", func(ctx SpecContext) {
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, "helm_values_yaml_parser", "initial commit")

			gotTestLine := false
			Expect(werfRender(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
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
