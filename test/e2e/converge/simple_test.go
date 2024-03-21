package e2e_converge_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("Simple converge", Label("e2e", "converge", "simple"), func() {
	var repoDirname string

	AfterEach(func() {
		utils.RunSucceedCommand(SuiteData.GetTestRepoPath(repoDirname), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
	})

	It("should succeed and deploy expected resources",
		func() {
			By("initializing")
			repoDirname = "repo0"
			setupEnv()

			By("state0: starting")
			{
				fixtureRelPath := "simple/state0"
				deployReportName := ".werf-deploy-report.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state0: execute converge")
				_, deployReport := werfProject.ConvergeWithReport(SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{})

				By("state0: check deploy report")
				Expect(deployReport.Release).To(Equal(werfProject.Release()))
				Expect(deployReport.Namespace).To(Equal(werfProject.Namespace()))
				Expect(deployReport.Revision).To(Equal(1))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state0: check deployed resources in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "test1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{"key1": "value1"}))
			}
		},
	)
})
