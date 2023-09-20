package e2e_converge_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/release"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("Complex converge", Label("e2e", "converge", "complex"), func() {
	var repoDirname string

	AfterEach(func() {
		utils.RunSucceedCommand(SuiteData.GetTestRepoPath(repoDirname), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
	})

	It("should complete and deploy expected resources",
		func() {
			By("initializing")
			repoDirname = "repo0"
			setupEnv()

			By("state0: starting")
			{
				fixtureRelPath := "complex/state0"
				deployReportName := ".werf-deploy-report.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state0: prepare namespace")
				werfProject.CreateNamespace()
				werfProject.CreateRegistryPullSecretFromDockerConfig()

				By("state0: execute converge")
				_, deployReport := werfProject.ConvergeWithReport(SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--set=app1.option=optionValue",
						},
					},
				})

				By("state0: check deploy report")
				Expect(deployReport.Release).To(Equal(werfProject.Release()))
				Expect(deployReport.Namespace).To(Equal(werfProject.Namespace()))
				Expect(deployReport.Revision).To(Equal(1))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state0: check deployed resources in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "app1-config", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"werfNamespace": werfProject.Namespace(),
					"werfEnv":       "test",
					"option":        "optionValue",
					"secretOption":  "secretOptionValue",
					"secretConfig":  "secretConfigContent",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)

				cm, err = kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "local-chart-config", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"werfEnv": "test",
					"option":  "optionValue",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)

				cm, err = kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "hello", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{"hello": "world"}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)

				deployment, err := kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), "app1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentAvailable(deployment)).To(BeTrue())
				checkServiceLabelsAndAnnos(deployment.Labels, deployment.Annotations, werfProject)

				deployment, err = kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), "app2", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentAvailable(deployment)).To(BeFalse())
				checkServiceLabelsAndAnnos(deployment.Labels, deployment.Annotations, werfProject)
			}

			By("state1: starting")
			{
				fixtureRelPath := "complex/state1"
				deployReportName := ".werf-deploy-report.json"

				By("state1: preparing test repo")
				SuiteData.UpdateTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state1: patch configmap in cluster, emulating manual changes by a user")
				_, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Patch(
					context.Background(),
					"app1-config",
					types.StrategicMergePatchType,
					[]byte(`{"data":{"option3": "setInClusterValue"}}`),
					metav1.PatchOptions{},
				)
				Expect(err).NotTo(HaveOccurred())

				By("state1: execute converge")
				_, deployReport := werfProject.ConvergeWithReport(SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--set=app1.option=optionValue",
						},
					},
				})

				By("state1: check deploy report")
				Expect(deployReport.Revision).To(Equal(2))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state1: check deployed configmap in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "app1-config", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"option2": "optionValue",
					"option3": "setInClusterValue",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)

				By("state1: check removed resources in cluster")
				resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(
					context.Background(), "local-chart-config", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(
					context.Background(), "hello", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(
					context.Background(), "app1", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(
					context.Background(), "app2", metav1.GetOptions{},
				))
			}

			By("state2: starting")
			{
				fixtureRelPath := "complex/state2"
				deployReportName := ".werf-deploy-report.json"

				By("state2: preparing test repo")
				SuiteData.UpdateTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state2: execute converge")
				_, deployReport := werfProject.ConvergeWithReport(SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
					},
				})

				By("state2: check deploy report")
				Expect(deployReport.Revision).To(Equal(3))
				Expect(deployReport.Status).To(Equal(release.StatusFailed))

				By("state2: check deployed deployment in cluster")
				deployment, err := kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), "app2", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentAvailable(deployment)).To(BeFalse())
				checkServiceLabelsAndAnnos(deployment.Labels, deployment.Annotations, werfProject)

				By("state2: check removed resources in cluster")
				resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(
					context.Background(), "app1-config", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(
					context.Background(), "local-chart-config", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(
					context.Background(), "hello", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(
					context.Background(), "app1", metav1.GetOptions{},
				))
				resourceShouldNotExist(kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(
					context.Background(), "app2", metav1.GetOptions{},
				))
			}
		},
	)
})

func checkServiceLabelsAndAnnos(labels, annotations map[string]string, werfProject *werf.Project) {
	Expect(labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"))

	Expect(annotations).To(HaveKeyWithValue("meta.helm.sh/release-name", werfProject.Release()))
	Expect(annotations).To(HaveKeyWithValue("meta.helm.sh/release-namespace", werfProject.Namespace()))
	Expect(annotations).To(HaveKeyWithValue("werf.io/version", "dev"))
	Expect(annotations).To(HaveKeyWithValue("project.werf.io/env", "test"))
}

func deploymentAvailable(deployment *appsv1.Deployment) bool {
	for _, cond := range deployment.Status.Conditions {
		if cond.Type == "Available" {
			return cond.Status == "True"
		}
	}

	return false
}

func resourceShouldNotExist(_ interface{}, err error) {
	if !apierrors.IsNotFound(err) {
		Expect(err).NotTo(HaveOccurred())
	}
}
