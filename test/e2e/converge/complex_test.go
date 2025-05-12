package e2e_converge_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/werf/3p-helm/pkg/release"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Complex converge", Label("e2e", "converge", "complex"), func() {
	var repoDirname string
	var werfProject *werf.Project

	crdsNames := []string{
		"crds-rootchart.example.org",
		"crds-subchart.example.org",
	}

	AfterEach(func() {
		utils.RunSucceedCommand(
			SuiteData.GetTestRepoPath(repoDirname),
			SuiteData.WerfBinPath,
			"dismiss",
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
				werfProject = werf.NewProject(
					SuiteData.WerfBinPath,
					SuiteData.GetTestRepoPath(repoDirname),
				)

				By("state0: prepare namespace")
				werfProject.CreateNamespace()
				werfProject.CreateRegistryPullSecretFromDockerConfig()

				By("state0: execute converge")
				convergeOutput, deployReport := werfProject.ConvergeWithReport(
					SuiteData.GetDeployReportPath(deployReportName),
					&werf.ConvergeWithReportOptions{
						CommonOptions: werf.CommonOptions{
							ExtraArgs: []string{
								"--set=added_via_set=added_via_set,overridden_via_set=overridden_via_set",
								"--set=subchart.added_via_set=added_via_set,subchart.overridden_via_set=overridden_via_set",
								"--set=added_via_set_list[0]=added_via_set,overridden_via_set_list[0]=overridden_via_set",
								"--set-string=added_via_set_string=added_via_set_string,overridden_via_set_string=overridden_via_set_string",
								"--values=.helm/values-extra.yaml",
								"--secret-values=.helm/secret-values-extra.yaml",
								"--add-annotation=added_via_add_annotation=added_via_add_annotation",
								"--add-label=added_via_add_label=added_via_add_label",
								"--set=disabledchart.enabled=false",
							},
						},
					},
				)

				By("state0: check deploy report")
				Expect(deployReport.Release).To(Equal(werfProject.Release()))
				Expect(deployReport.Namespace).To(Equal(werfProject.Namespace()))
				Expect(deployReport.Revision).To(Equal(1))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state0: check configmap config-rootchart in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "config-rootchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"werf_namespace": werfProject.Namespace(),
					"werf_env":       "test",

					"chart_name":                   "rootchart",
					"chart_version":                "1.0.0",
					"chart_app_version":            "0.1.0",
					"chart_deprecated":             "false",
					"chart_icon":                   "myicon",
					"chart_description":            "mydescription",
					"chart_home":                   "myhome",
					"chart_first_source":           "mysource",
					"chart_first_keyword":          "mykeyword",
					"chart_first_annotation":       "myannovalue",
					"chart_first_maintainer_name":  "myname",
					"chart_first_maintainer_email": "myemail",
					"chart_first_maintainer_url":   "myurl",

					"release_is_install": "true",
					"release_is_upgrade": "false",
					"release_name":       werfProject.Release(),
					"release_namespace":  werfProject.Namespace(),
					"release_revision":   "1",

					"template_base_path": "rootchart/templates",
					"template_name":      "rootchart/templates/configmap.yaml",

					"capabilities_kube_version_major":  "1",
					"capabilities_api_versions_has_v1": "true",

					"global_preserved":                   "preserved",
					"preserved":                          "preserved",
					"added_via_set":                      "added_via_set",
					"added_via_set_string":               "added_via_set_string",
					"added_via_values":                   "added_via_values",
					"added_via_secret_values":            "added_via_secret_values",
					"added_via_secret_values_extra":      "added_via_secret_values_extra",
					"added_via_set_list":                 "added_via_set",
					"overridden_via_set":                 "overridden_via_set",
					"overridden_via_set_string":          "overridden_via_set_string",
					"overridden_via_values":              "overridden_via_values",
					"overridden_via_secret_values":       "overridden_via_secret_values",
					"overridden_via_secret_values_extra": "overridden_via_secret_values_extra",

					"preserved_list":             "preserved",
					"overridden_via_set_list":    "overridden_via_set",
					"overridden_via_values_list": "overridden_via_values",

					"import_preserved_via_import": "preserved_via_import",
					"import_added_via_import":     "added_via_import",

					"secret_config": "secretConfigContent",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)
				checkGlobalLabelsAndAnnos(cm.Labels, cm.Annotations)

				By("state0: check configmap config-subchart in cluster")
				cm, err = kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "config-subchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"werf_namespace": werfProject.Namespace(),
					"werf_env":       "test",

					"chart_name":    "subchart",
					"chart_version": "0.1.0",

					"template_base_path": "rootchart/charts/subchart/templates",
					"template_name":      "rootchart/charts/subchart/templates/configmap.yaml",

					"global_preserved":                  "preserved",
					"preserved":                         "preserved",
					"added_via_set":                     "added_via_set",
					"added_via_parent_values":           "added_via_parent_values",
					"overridden_via_set":                "overridden_via_set",
					"overridden_via_parent_values":      "overridden_via_parent_values",
					"overridden_via_parent_values_list": "overridden_via_parent_values",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)
				checkGlobalLabelsAndAnnos(cm.Labels, cm.Annotations)

				for _, deploymentName := range []string{
					"deployment-rootchart",
					"deployment-subchart",
					"deployment-subsubchart",
					"hook-rootchart",
				} {
					By("state0: check deployment \"" + deploymentName + "\" in cluster")
					deployment, err := kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), deploymentName, metav1.GetOptions{})
					Expect(err).NotTo(HaveOccurred())
					Expect(deploymentAvailable(deployment)).To(BeTrue())
					checkServiceLabelsAndAnnos(deployment.Labels, deployment.Annotations, werfProject)
					checkGlobalLabelsAndAnnos(cm.Labels, cm.Annotations)
				}

				for _, configMapName := range []string{
					"config-aliasedchart",
					"hello",
				} {
					By("state0: check configmap \"" + configMapName + "\" in cluster")
					_, err = kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), configMapName, metav1.GetOptions{})
					Expect(err).NotTo(HaveOccurred())
					checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)
					checkGlobalLabelsAndAnnos(cm.Labels, cm.Annotations)
				}

				for _, configMapName := range []string{
					"config-disabledchart",
					"not-deployed-because-in-helm-ignore",
				} {
					By("state0: ensure configmap \"" + configMapName + "\" is absent in cluster")
					resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), configMapName, metav1.GetOptions{}))
				}

				for _, serviceName := range []string{
					"service-rootchart",
					"service-subchart",
					"service-hook-rootchart",
				} {
					By("state0: check service \"" + serviceName + "\" in cluster")
					_, err = kube.Client.CoreV1().Services(werfProject.Namespace()).Get(context.Background(), serviceName, metav1.GetOptions{})
					Expect(err).NotTo(HaveOccurred())
					checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)
					checkGlobalLabelsAndAnnos(cm.Labels, cm.Annotations)
				}

				By("state0: ensure job \"hook-subchart\" is absent in cluster")
				resourceShouldNotExist(kube.Client.BatchV1().Jobs(werfProject.Namespace()).Get(context.Background(), "hook-subchart", metav1.GetOptions{}))

				By("state0: check crd \"crds-rootchart\" in cluster")
				_, err = kube.DynamicClient.Resource(schema.GroupVersionResource{
					Group:    "example.org",
					Version:  "v1",
					Resource: "crds-rootchart",
				}).Namespace(werfProject.Namespace()).Get(context.Background(), "cr-rootchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				By("state0: check crd \"crds-subchart\" in cluster")
				_, err = kube.DynamicClient.Resource(schema.GroupVersionResource{
					Group:    "example.org",
					Version:  "v1",
					Resource: "crds-subchart",
				}).Namespace(werfProject.Namespace()).Get(context.Background(), "cr-subchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				By("state0: ensure notes rendered")
				Expect(convergeOutput).To(ContainSubstring("mynotes"))

				By("state0: ensure hooks executed")
				Expect(convergeOutput).To(ContainSubstring("hook-subchart completed"))
			}

			By("state1: starting")
			{
				fixtureRelPath := "complex/state1"
				deployReportName := ".werf-deploy-report.json"

				By("state1: preparing test repo")
				SuiteData.UpdateTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state1: simulate manual user changes to the configmap \"config-rootchart\" by `kubectl edit`-like patching it in the cluster")
				_, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Patch(
					context.Background(),
					"config-rootchart",
					types.StrategicMergePatchType,
					[]byte(`{
"data":
	{
		"reset_after_manual_changes_in_cluster": "will_be_overridden", "removed_after_manual_changes_in_cluster": "will_be_removed"
	}
}`),
					metav1.PatchOptions{
						FieldManager: "kubectl-edit",
					},
				)
				Expect(err).NotTo(HaveOccurred())

				By("state1: execute converge")
				_, deployReport := werfProject.ConvergeWithReport(SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{})

				By("state1: check deploy report")
				Expect(deployReport.Revision).To(Equal(2))
				Expect(deployReport.Status).To(Equal(release.StatusDeployed))

				By("state1: check configmap \"config-rootchart\" in the cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "config-rootchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"release_is_install": "false",
					"release_is_upgrade": "true",
					"release_revision":   "2",

					"reset_after_manual_changes_in_cluster": "reset_after_manual_changes_in_cluster",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)

				By("state1: check deployment \"hook-rootchart\" in cluster")
				_, err = kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), "hook-rootchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				for _, crdName := range crdsNames {
					By("state1: check crd \"" + crdName + "\" in cluster")
					_, err = kube.DynamicClient.Resource(schema.GroupVersionResource{
						Group:    "apiextensions.k8s.io",
						Version:  "v1",
						Resource: "customresourcedefinitions",
					}).Get(context.Background(), crdName, metav1.GetOptions{})
					Expect(err).NotTo(HaveOccurred())
				}

				for _, deploymentName := range []string{
					"deployment-rootchart",
					"deployment-subchart",
					"deployment-subsubchart",
				} {
					By("state1: ensure deployment \"" + deploymentName + "\" is absent in cluster")
					resourceShouldNotExist(kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), deploymentName, metav1.GetOptions{}))
				}

				for _, configMapName := range []string{
					"config-subchart",
					"config-aliasedchart",
					"hello",
				} {
					By("state1: ensure configmap \"" + configMapName + "\" is absent in cluster")
					resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), configMapName, metav1.GetOptions{}))
				}

				for _, serviceName := range []string{
					"service-rootchart",
					"service-subchart",
					"service-hook-rootchart",
				} {
					By("state1: ensure service \"" + serviceName + "\" is absent in cluster")
					resourceShouldNotExist(kube.Client.CoreV1().Services(werfProject.Namespace()).Get(context.Background(), serviceName, metav1.GetOptions{}))
				}

				By("state1: ensure crd \"CRDRootchart\" is absent in cluster")
				resourceShouldNotExist(kube.DynamicClient.Resource(schema.GroupVersionResource{
					Group:    "example.org",
					Version:  "v1",
					Resource: "crds-rootchart",
				}).Namespace(werfProject.Namespace()).Get(context.Background(), "cr-rootchart", metav1.GetOptions{}))

				By("state1: ensure crd \"CRDSubchart\" is absent in cluster")
				resourceShouldNotExist(kube.DynamicClient.Resource(schema.GroupVersionResource{
					Group:    "example.org",
					Version:  "v1",
					Resource: "crds-subchart",
				}).Namespace(werfProject.Namespace()).Get(context.Background(), "cr-subchart", metav1.GetOptions{}))
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
						ExtraArgs: []string{
							"--auto-rollback",
						},
					},
				})

				By("state2: check deploy report")
				Expect(deployReport.Revision).To(Equal(3))
				Expect(deployReport.Status).To(Equal(release.StatusFailed))

				By("state2: check configmap \"config-rootchart\" in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "config-rootchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{
					"release_is_install": "false",
					"release_is_upgrade": "true",
					"release_revision":   "2",

					"reset_after_manual_changes_in_cluster": "reset_after_manual_changes_in_cluster",
				}))
				checkServiceLabelsAndAnnos(cm.Labels, cm.Annotations, werfProject)

				By("state2: ensure deployment \"deployment-rootchart\" is absent in cluster")
				resourceShouldNotExist(kube.Client.AppsV1().Deployments(werfProject.Namespace()).Get(context.Background(), "deployment-rootchart", metav1.GetOptions{}))
			}

			By("state3: starting")
			{
				fixtureRelPath := "complex/state3"
				deployReportName := ".werf-deploy-report.json"

				By("state3: preparing test repo")
				SuiteData.UpdateTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state3: execute converge")
				_, deployReport := werfProject.ConvergeWithReport(SuiteData.GetDeployReportPath(deployReportName), &werf.ConvergeWithReportOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
					},
				})

				By("state3: check deploy report")
				Expect(deployReport.Revision).To(Equal(5))
				Expect(deployReport.Status).To(Equal(release.StatusFailed))

				By("state3: check job \"hook-rootchart\" in cluster")
				job, err := kube.Client.BatchV1().Jobs(werfProject.Namespace()).Get(context.Background(), "hook-rootchart", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				checkServiceLabelsAndAnnos(job.Labels, job.Annotations, werfProject)

				By("state3: ensure configmap \"config-rootchart\" is absent in cluster")
				resourceShouldNotExist(kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "config-rootchart", metav1.GetOptions{}))
			}
		},
	)
})

func checkServiceLabelsAndAnnos(labels, annotations map[string]string, werfProject *werf.Project) {
	if _, isHook := annotations["helm.sh/hook"]; !isHook {
		Expect(labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"))

		Expect(annotations).To(HaveKeyWithValue("meta.helm.sh/release-name", werfProject.Release()))
		Expect(annotations).To(HaveKeyWithValue("meta.helm.sh/release-namespace", werfProject.Namespace()))
	}

	Expect(annotations).To(HaveKeyWithValue("werf.io/version", "dev"))
	Expect(annotations).To(HaveKeyWithValue("project.werf.io/env", "test"))
}

func checkGlobalLabelsAndAnnos(labels, annotations map[string]string) {
	Expect(labels).To(HaveKeyWithValue("added_via_add_label", "added_via_add_label"))

	Expect(annotations).To(HaveKeyWithValue("added_via_add_annotation", "added_via_add_annotation"))
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
