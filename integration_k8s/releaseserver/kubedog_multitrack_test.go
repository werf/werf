package releaseserver_test

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func releaseResourcesStatusProgressLine(outputLine string) string {
	prefix := "│ │ │ "
	if strings.HasPrefix(outputLine, prefix) {
		return strings.Trim(outputLine, prefix)
	}
	return ""
}

type DeploymentState struct {
	Deployment       string
	Replicas         string
	ReplicasCurrent  string
	Available        string
	AvailableCurrent string
	UpToDate         string
	UpToDateCurrent  string

	StatusProgressLine string
}

func mydeploy1State(statusProgressLine string) *DeploymentState {
	fields := strings.Fields(statusProgressLine)

	if len(fields) == 4 && fields[0] == "mydeploy1" {
		ds := &DeploymentState{
			Deployment:         fields[0],
			Replicas:           fields[1],
			Available:          fields[2],
			UpToDate:           fields[3],
			StatusProgressLine: statusProgressLine,
		}

		if parts := strings.Split(ds.Replicas, "->"); len(parts) > 1 {
			ds.ReplicasCurrent = parts[1]
		} else {
			ds.ReplicasCurrent = ds.Replicas
		}
		if parts := strings.Split(ds.Available, "->"); len(parts) > 1 {
			ds.AvailableCurrent = parts[1]
		} else {
			ds.AvailableCurrent = ds.Available
		}
		if parts := strings.Split(ds.UpToDate, "->"); len(parts) > 1 {
			ds.UpToDateCurrent = parts[1]
		} else {
			ds.UpToDateCurrent = ds.UpToDate
		}

		return ds
	}

	return nil
}

func unknownDeploymentStateForbidden(ds *DeploymentState) {
	Expect(ds.Replicas).ShouldNot(Equal("-"), fmt.Sprintf("Unknown deploy/%s REPLICAS should not be reported", ds.Deployment))
	Expect(ds.Available).ShouldNot(Equal("-"), fmt.Sprintf("Unknown deploy/%s AVAILABLE should not be reported", ds.Deployment))
	Expect(ds.UpToDate).ShouldNot(Equal("-"), fmt.Sprintf("Unknown deploy/%s UP-TO-DATE should not be reported", ds.Deployment))
}

var _ = Describe("Kubedog multitrack — werf's kubernetes resources tracker", func() {
	Context("when chart contains valid resource", func() {
		AfterEach(func() {
			utils.RunCommand("kubedog_multitrack_app1", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should report Deployment is ready before werf exit", func() {
			gotDeploymentReadyLine := false

			Expect(werfDeploy("kubedog_multitrack_app1", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy1 := mydeploy1State(statusProgressLine); mydeploy1 != nil {
							unknownDeploymentStateForbidden(mydeploy1)

							if mydeploy1.UpToDateCurrent == "2" && mydeploy1.AvailableCurrent == "2" && mydeploy1.ReplicasCurrent == "2/2" {
								gotDeploymentReadyLine = true
							}
						}
					}
				},
			})).Should(Succeed())

			Expect(gotDeploymentReadyLine).Should(BeTrue())
		})
	})

	Context("when chart contains resource with invalid docker image", func() {
		AfterEach(func() {
			utils.RunCommand("kubedog_multitrack_app2", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should report ImagePullBackoff occured in Deployment and werf should fail", func() {
			gotImagePullBackoffLine := false
			gotAllowedErrorsWarning := false
			gotAllowedErrorsExceeded := false

			Expect(werfDeploy("kubedog_multitrack_app2", liveexec.ExecCommandOptions{
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

					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy1 := mydeploy1State(statusProgressLine); mydeploy1 != nil {
							unknownDeploymentStateForbidden(mydeploy1)
						}
					}
				},
			})).Should(MatchError("exit code 1"))

			Expect(gotImagePullBackoffLine).Should(BeTrue())
			Expect(gotAllowedErrorsWarning).Should(BeTrue())
			Expect(gotAllowedErrorsExceeded).Should(BeTrue())
		})
	})
})
