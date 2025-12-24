package e2e_converge_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/3p-helm/pkg/release"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple converge", Label("e2e", "converge", "simple"), func() {
	var repoDirname string
	var werfProject *werf.Project

	AfterEach(func(ctx SpecContext) {
		utils.RunSucceedCommand(
			ctx,
			SuiteData.GetTestRepoPath(repoDirname),
			SuiteData.WerfBinPath,
			"dismiss",
			"--with-namespace",
		)

		werfProject.KubeCtl(ctx, &werf.KubeCtlOptions{
			werf.CommonOptions{
				ExtraArgs: []string{
					"delete",
					"namespace",
					"--ignore-not-found",
					werfProject.Namespace(ctx),
				},
			},
		})
	})

	It("should succeed and deploy expected resources",
		func(ctx SpecContext) {
			By("initializing")
			repoDirname = "repo0"
			setupEnv()

			By("state0: starting")
			{
				fixtureRelPath := "simple/state0"
				deployReportName := ".werf-deploy-report.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)
				werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)

				By("state0: execute converge")
				_, deployReport := reportProject.ConvergeWithReport(ctx, SuiteData.GetDeployReportPath(deployReportName), &werf.WithReportOptions{})

				By("state0: check deploy report")
				Expect(deployReport.Release).To(Equal(werfProject.Release(ctx)))
				Expect(deployReport.Namespace).To(Equal(werfProject.Namespace(ctx)))
				Expect(deployReport.Revision).To(Equal(1))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state0: check deployed resources in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace(ctx)).Get(ctx, "test1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{"key1": "value1"}))
			}
		},
	)

	It("should succeed and deploy expected resources with using build report",
		func(ctx SpecContext) {
			By("initializing")
			repoDirname = "repo0"
			setupEnv()

			By("state0: starting")
			{
				fixtureRelPath := "simple/state0"
				deployReportName := ".werf-deploy-report.json"
				buildReportName := "report0.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)
				werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state0: building images")
				buildOut, _ := werfProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

				By("state0: execute converge")
				convergeOut, deployReport := werfProject.ConvergeWithReport(ctx, SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{"--use-build-report", "--build-report-path", SuiteData.GetBuildReportPath(buildReportName)},
					},
				})

				By("state0: check converge output")
				Expect(convergeOut).To(ContainSubstring(fmt.Sprintf("Avoid buildibg because of using build report: %s", SuiteData.GetBuildReportPath(buildReportName))))

				By("state0: check deploy report")
				Expect(deployReport.Release).To(Equal(werfProject.Release(ctx)))
				Expect(deployReport.Namespace).To(Equal(werfProject.Namespace(ctx)))
				Expect(deployReport.Revision).To(Equal(1))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state0: check deployed resources in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace(ctx)).Get(ctx, "test1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{"key1": "value1"}))
			}
		},
	)
})
