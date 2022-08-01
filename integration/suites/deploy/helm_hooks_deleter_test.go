package deploy_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

var _ = Describe("Helm hooks deleter", func() {
	Context("when installing chart with post-install Job hook and hook-succeeded delete policy", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should delete hook when hook succeeded and wait till it is deleted without timeout https://github.com/werf/werf/issues/1885", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "helm_hooks_deleter_app1", "initial commit")

			gotDeletingHookLine := false
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.HasPrefix(line, "│ NOTICE Will not delete Job/migrate: resource does not belong to the helm release")).ShouldNot(BeTrue(), fmt.Sprintf("Got unexpected output line: %v", line))

					if strings.Contains(line, "Waiting for resources elimination: jobs/migrate") {
						gotDeletingHookLine = true
					}
				},
			})).Should(Succeed())

			Expect(gotDeletingHookLine).Should(BeTrue())
		})
	})
})
