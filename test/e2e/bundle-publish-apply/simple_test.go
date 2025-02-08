package e2e_bundle_publish_apply_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/3p-helm/pkg/release"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple bundle publish/apply", Label("e2e", "bundle-publish-apply", "simple"), func() {
	var repoDirname string
	var werfProject *werf.Project

	AfterEach(func() {
		utils.RunSucceedCommand(
			SuiteData.GetTestRepoPath(repoDirname),
			SuiteData.WerfBinPath,
			"dismiss",
			"--release",
			werfProject.Release(),
			"--namespace",
			werfProject.Namespace(),
			"--with-namespace",
		)

		werfProject.KubeCtl(&werf.KubeCtlOptions{
			werf.CommonOptions{
				ExtraArgs: []string{
					"delete",
					"namespace",
					"--ignore-not-found",
					werfProject.Namespace(),
				},
			},
		})
	})

	It("should succeed and deploy expected resources",
		func() {
			By("initializing")
			repoDirname = "repo0"
			setupEnv()
			contRuntime, err := contback.NewContainerBackend("vanilla-docker")
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("state0: starting")
			{
				fixtureRelPath := "simple/state0"
				buildReportName := ".werf-build-report.json"
				deployReportName := ".werf-deploy-report.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)
				werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state0: execute bundle publish")
				_, buildReport := werfProject.BundlePublishWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)

				By(`state0: checking "dockerfile" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /file",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-run",
				)

				By("state0: execute bundle apply")
				_, deployReport := werfProject.BundleApplyWithReport(werfProject.Release(), werfProject.Namespace(), SuiteData.GetDeployReportPath(deployReportName), nil)

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
