package releaseserver_test

import (
	"fmt"
	"strings"

	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/integration/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helm hooks deleter", func() {
	Context("when installing chart with post-install Job hook and hook-succeeded delete policy", func() {
		AfterEach(func() {
			utils.RunCommand("helm_hooks_deleter_app1", werfBinPath, "dismiss", "--with-namespace")
		})

		It("should delete hook when hook succeeded and wait till it is deleted without timeout https://github.com/werf/werf/issues/1885", func() {
			gotDeletingHookLine := false

			Expect(werfDeploy("helm_hooks_deleter_app1", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.HasPrefix(line, "│ NOTICE Will not delete Job/migrate: resource does not belong to the helm release")).ShouldNot(BeTrue(), fmt.Sprintf("Got unexpected output line: %v", line))

					if strings.HasPrefix(line, "Deleting Job/migrate of release") {
						gotDeletingHookLine = true
					}
				},
			})).Should(Succeed())

			Expect(gotDeletingHookLine).Should(BeTrue())
		})
	})
})
