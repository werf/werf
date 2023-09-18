package deploy_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

func releaseResourcesStatusProgressLine(outputLine string) string {
	prefix := "│ │ "
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

func deployState(statusProgressLine, deployName string) *DeploymentState {
	fields := strings.Fields(statusProgressLine)

	if len(fields) == 4 && fields[0] == deployName {
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

func isTooManyProbesTriggered(line, probeName string, maxAllowed int) bool {
	var isTooManyProbesTriggered bool

	numberRegex := regexp.MustCompile("[0-9]+")

	if strings.Contains(line, fmt.Sprintf(`Count of triggered %s probes:`, probeName)) {
		if triggeredProbes, err := strconv.Atoi(numberRegex.FindAllString(line, 1)[0]); err != nil {
			Fail(err.Error())
		} else if triggeredProbes > maxAllowed {
			isTooManyProbesTriggered = true
		}
	}
	return isTooManyProbesTriggered
}

var _ = Describe("Kubedog multitrack — werf's kubernetes resources tracker", func() {
	Context("when chart contains valid resource", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should report Deployment is ready before werf exit", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "kubedog_multitrack_app1", "initial commit")

			gotDeploymentReadyLine := false

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					line = stripansi.Strip(line)
					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy1 := deployState(statusProgressLine, "mydeploy1"); mydeploy1 != nil {
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
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should report ImagePullBackoff occurred in Deployment and werf should fail", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "kubedog_multitrack_app2", "initial commit")

			gotImagePullBackoffLine := false
			gotAllowedErrorsWarning := false
			gotAllowedErrorsExceeded := false

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, `1/1 allowed errors occurred for deploy/mydeploy1: continue tracking`) {
						gotAllowedErrorsWarning = true
					}
					if strings.Contains(line, `Allowed failures count for deploy/mydeploy1 exceeded 1 errors: stop tracking immediately!`) {
						gotAllowedErrorsExceeded = true
					}
					if strings.Contains(line, "deploy/mydeploy1 ERROR:") && strings.Contains(line, `ImagePullBackOff: Back-off pulling image "ubuntu:18.03"`) {
						gotImagePullBackoffLine = true
					}

					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy1 := deployState(statusProgressLine, "mydeploy1"); mydeploy1 != nil {
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

	Context("when chart contains resource with succeeding probes", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should report Deployment is ready before werf exit", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "kubedog_multitrack_app3", "initial commit")

			var gotDeploymentReadyLine bool

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					line = stripansi.Strip(line)
					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy3 := deployState(statusProgressLine, "mydeploy3"); mydeploy3 != nil {
							unknownDeploymentStateForbidden(mydeploy3)

							if mydeploy3.UpToDateCurrent == "2" && mydeploy3.AvailableCurrent == "2" && mydeploy3.ReplicasCurrent == "2/2" {
								gotDeploymentReadyLine = true
							}
						}
					}
				},
			})).Should(Succeed())

			Expect(gotDeploymentReadyLine).Should(BeTrue())
		})
	})

	Context("when chart contains resource with failing startup probe", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should report container killed by startup probe and werf should fail", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "kubedog_multitrack_app4", "initial commit")

			var gotKilledByStartupProbe, gotAllowedErrorsExceeded bool

			startupFailRegex := regexp.MustCompile("deploy/mydeploy4 ERROR: po/mydeploy4-[a-z0-9]+-[a-z0-9]+ container/main: Killing: Container main failed startup probe, will be restarted")

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if startupFailRegex.MatchString(line) {
						gotKilledByStartupProbe = true
					}
					if strings.Contains(line, `Allowed failures count for deploy/mydeploy4 exceeded 0 errors: stop tracking immediately!`) {
						gotAllowedErrorsExceeded = true
					}

					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy4 := deployState(statusProgressLine, "mydeploy4"); mydeploy4 != nil {
							unknownDeploymentStateForbidden(mydeploy4)
						}
					}
				},
			})).Should(MatchError("exit code 1"))

			Expect(gotKilledByStartupProbe).Should(BeTrue())
			Expect(gotAllowedErrorsExceeded).Should(BeTrue())
		})
	})

	Context("when chart contains resource with failing readiness probe", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should report that the container readiness probe failed multiple times and werf should fail", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "kubedog_multitrack_app5", "initial commit")

			var gotStoppedByReadinessProbe, gotAllowedErrorsExceeded, gotTooManyProbesTriggered bool

			readinessFailRegex := regexp.MustCompile("deploy/mydeploy5 ERROR: po/mydeploy5-[a-z0-9]+-[a-z0-9]+ container/main: Unhealthy: Readiness probe failed")

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if readinessFailRegex.MatchString(line) {
						gotStoppedByReadinessProbe = true
					}
					if strings.Contains(line, `Allowed failures count for deploy/mydeploy5 exceeded 2 errors: stop tracking immediately!`) {
						gotAllowedErrorsExceeded = true
					}
					if isTooManyProbesTriggered(line, "readiness", 8) {
						gotTooManyProbesTriggered = true
					}

					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy5 := deployState(statusProgressLine, "mydeploy5"); mydeploy5 != nil {
							unknownDeploymentStateForbidden(mydeploy5)
						}
					}
				},
			})).Should(MatchError("exit code 1"))

			Expect(gotStoppedByReadinessProbe).Should(BeTrue())
			Expect(gotAllowedErrorsExceeded).Should(BeTrue())
			Expect(gotTooManyProbesTriggered).Should(BeFalse())
		})
	})

	Context("when chart contains resource with failing liveness probe", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should report that the container liveness probe failed and werf should fail", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "kubedog_multitrack_app6", "initial commit")

			var gotKilledByLivenessProbe, gotAllowedErrorsExceeded bool

			livenessFailRegex := regexp.MustCompile("deploy/mydeploy6 ERROR: po/mydeploy6-[a-z0-9]+-[a-z0-9]+ container/main: Killing: Container main failed liveness probe, will be restarted")

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if livenessFailRegex.MatchString(line) {
						gotKilledByLivenessProbe = true
					}
					if strings.Contains(line, `Allowed failures count for deploy/mydeploy6 exceeded 0 errors: stop tracking immediately!`) {
						gotAllowedErrorsExceeded = true
					}

					if statusProgressLine := releaseResourcesStatusProgressLine(line); statusProgressLine != "" {
						if mydeploy6 := deployState(statusProgressLine, "mydeploy6"); mydeploy6 != nil {
							unknownDeploymentStateForbidden(mydeploy6)
						}
					}
				},
			})).Should(MatchError("exit code 1"))

			Expect(gotKilledByLivenessProbe).Should(BeTrue())
			Expect(gotAllowedErrorsExceeded).Should(BeTrue())
		})
	})
})
