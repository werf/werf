// +build integration_k8s

package deploytracking

import (
	"strings"

	"github.com/flant/werf/integration/utils/werfexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func resourceStateLine(outputLine string) string {
	fields := strings.Fields(outputLine)
	if len(fields) >= 4 {
		return strings.Join(fields[len(fields)-4:len(fields)], " ")
	}
	return ""
}

func unknownMydeploy1ResourceStateForbidden(outputLine string) {
	resourceState := resourceStateLine(outputLine)
	Ω(resourceState).ShouldNot(Equal("mydeploy1 - - -"), "Unknown mydeploy1 state should not be reported")
}

var _ = Describe("Kubedog multitrack — werf's kubernetes resources tracker", func() {
	Context("when chart contains valid resource", func() {
		AfterEach(func() {
			werfDismiss("app1", werfexec.CommandOptions{})
		})

		It("should report Deployment is ready before werf exit", func(done Done) {
			gotDeploymentReadyLine := false

			Ω(werfDeploy("app1", werfexec.CommandOptions{
				OutputLineHandler: func(line string) {
					resourceState := resourceStateLine(line)
					if resourceState == "mydeploy1 2/2 2 2" {
						gotDeploymentReadyLine = true
					}

					unknownMydeploy1ResourceStateForbidden(line)
				},
			})).Should(Succeed())

			Ω(gotDeploymentReadyLine).Should(BeTrue())

			close(done)
		}, 120)
	})

	Context("when chart contains resource with invalid docker image", func() {
		AfterEach(func() {
			werfDismiss("app2", werfexec.CommandOptions{})
		})

		It("should report ImagePullBackoff occured in Deployment and werf should fail", func(done Done) {
			gotImagePullBackoffLine := false
			gotAllowedErrorsWarning := false
			gotAllowedErrorsExceeded := false

			Ω(werfDeploy("app2", werfexec.CommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, `1/1 allowed errors occurred for deploy/mydeploy1: continue tracking`) != -1 {
						gotAllowedErrorsWarning = true
					}
					if strings.Index(line, `Allowed failures count for deploy/mydeploy1 exceeded 1 errors: stop tracking immediately!`) != -1 {
						gotAllowedErrorsExceeded = true
					}
					if strings.Index(line, "deploy/mydeploy1 ERROR:") != -1 && strings.HasSuffix(line, `ImagePullBackOff: Back-off pulling image "ubuntu:18.03"`) {
						gotImagePullBackoffLine = true
					}

					unknownMydeploy1ResourceStateForbidden(line)
				},
			})).Should(MatchError("exit code 1"))

			Ω(gotImagePullBackoffLine).Should(BeTrue())
			Ω(gotAllowedErrorsWarning).Should(BeTrue())
			Ω(gotAllowedErrorsExceeded).Should(BeTrue())

			close(done)
		}, 120)
	})
})

func werfDeploy(dir string, opts werfexec.CommandOptions) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, "deploy", "--env", "dev")
}

func werfDismiss(dir string, opts werfexec.CommandOptions) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, "dismiss", "--env", "dev", "--with-namespace")
}
