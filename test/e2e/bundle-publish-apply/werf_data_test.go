package e2e_bundle_publish_apply_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	helmregistry "github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/test/pkg/suite_init"
	"github.com/werf/werf/v3/test/pkg/utils"
	"github.com/werf/werf/v3/test/pkg/werf"
)

var publishInputsArgs = []string{
	"--values=inputs/values.yaml",
	"--secret-values=inputs/secret-values.yaml",
	"--set=fromSet=from-set",
	"--set-file=fromSetFile=inputs/setfile.txt",
}

var _ = Describe("Bundle publish/apply with .werf-data", Label("e2e", "bundle-publish-apply", "werf-data"), func() {
	var werfProject *werf.Project
	var applyProject *werf.Project

	BeforeEach(func(ctx SpecContext) {
		setupEnv()

		SuiteData.InitTestRepo(ctx, "app", "werf-data/app")
		werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("app"))

		applyDir := SuiteData.GetTestRepoPath("apply")
		Expect(os.MkdirAll(applyDir, os.ModePerm)).To(Succeed())
		applyProject = werf.NewProject(SuiteData.WerfBinPath, applyDir)
	})

	It("should apply to the recorded release and namespace and require them for another env",
		func(ctx SpecContext) {
			clientFactory := werf.NewKubeClientFactory(ctx)
			release, namespace := werfProject.Release(ctx), werfProject.Namespace(ctx)

			By("publishing the bundle")
			werfProject.BundlePublish(ctx, nil)

			By("applying the bundle without --release and --namespace")
			dismissOnCleanup(werfProject, release, namespace)
			applyOut := applyProject.RunCommand(ctx, []string{"bundle", "apply", "--log-terminal-width", "1000"}, werf.CommonOptions{})
			Expect(applyOut).To(ContainSubstring(fmt.Sprintf(`Using release name %q and release namespace %q recorded in .werf-data`, release, namespace)))

			By("checking the release in the recorded namespace")
			releaseGet := getRelease(ctx, applyProject, release, namespace)
			Expect(releaseGet.Release.Revision).To(Equal(1))
			cm, err := clientFactory.Static().CoreV1().ConfigMaps(namespace).Get(ctx, "app-cfg", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cm.Data).To(HaveKeyWithValue("werfEnv", "test"))

			By("applying the bundle with another env")
			mismatchErr := `release defaults recorded for env "test" are not used with --env "other": pass`
			Expect(applyProject.RunCommand(ctx, []string{"bundle", "apply", "--env", "other"}, werf.CommonOptions{ShouldFail: true})).
				To(ContainSubstring(mismatchErr + " --release and --namespace"))
			Expect(applyProject.RunCommand(ctx, []string{"bundle", "apply", "--env", "other", "--release", release}, werf.CommonOptions{ShouldFail: true})).
				To(ContainSubstring(mismatchErr + " --namespace"))
		},
	)

	It("should deploy the publish-time inputs as converge does and let the apply-time inputs win",
		func(ctx SpecContext) {
			clientFactory := werf.NewKubeClientFactory(ctx)
			release, namespace := werfProject.Release(ctx), werfProject.Namespace(ctx)
			convergeRelease := SuiteData.ProjectName + "-converge"
			setSecretKeyEnv(werfProject)

			By("publishing the bundle")
			werfProject.BundlePublish(ctx, &werf.BundlePublishOptions{
				CommonOptions: werf.CommonOptions{ExtraArgs: publishInputsArgs},
			})

			By("applying the bundle from another directory")
			dismissOnCleanup(werfProject, release, namespace)
			applyProject.BundleApply(ctx, release, namespace, nil)

			By("converging the same inputs into another release")
			dismissOnCleanup(werfProject, convergeRelease, convergeRelease)
			werfProject.Converge(ctx, &werf.ConvergeOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: append([]string{"--release", convergeRelease, "--namespace", convergeRelease}, publishInputsArgs...),
				},
			})

			By("comparing the releases")
			Expect(getRelease(ctx, applyProject, release, namespace).Values).To(Equal(getRelease(ctx, applyProject, convergeRelease, convergeRelease).Values))
			cm, err := clientFactory.Static().CoreV1().ConfigMaps(namespace).Get(ctx, "app-cfg", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cm.Data).To(Equal(map[string]string{
				"fromValues":       "from-values-file",
				"fromSecretValues": "from-secret-values-file",
				"fromSet":          "from-set",
				"fromSetFile":      "from-set-file",
				"greeting":         "chart",
				"werfEnv":          "test",
			}))

			By("applying the bundle with apply-time inputs")
			applyTimeArgs := []string{"--set=fromSet=from-apply-set"}
			applyProject.BundleApply(ctx, release, namespace, &werf.BundleApplyOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: append([]string{"--values=" + filepath.Join(werfProject.GitRepoPath, "inputs/apply-values.yaml")}, applyTimeArgs...),
				},
			})

			By("converging the same inputs into another release")
			werfProject.Converge(ctx, &werf.ConvergeOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: append(append([]string{"--release", convergeRelease, "--namespace", convergeRelease}, publishInputsArgs...), append([]string{"--values=inputs/apply-values.yaml"}, applyTimeArgs...)...),
				},
			})

			By("comparing the releases")
			Expect(getRelease(ctx, applyProject, release, namespace).Values).To(Equal(getRelease(ctx, applyProject, convergeRelease, convergeRelease).Values))
			cm, err = clientFactory.Static().CoreV1().ConfigMaps(namespace).Get(ctx, "app-cfg", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cm.Data).To(Equal(map[string]string{
				"fromValues":       "from-apply-values-file",
				"fromSecretValues": "from-secret-values-file",
				"fromSet":          "from-apply-set",
				"fromSetFile":      "from-set-file",
				"greeting":         "chart",
				"werfEnv":          "test",
			}))
		},
	)

	It("should apply a saved plan of the bundle",
		func(ctx SpecContext) {
			clientFactory := werf.NewKubeClientFactory(ctx)
			release, namespace := werfProject.Release(ctx), werfProject.Namespace(ctx)
			planPath := filepath.Join(applyProject.GitRepoPath, "bundle.plan.gz")

			By("publishing the bundle")
			werfProject.BundlePublish(ctx, nil)

			By("saving the plan")
			applyProject.RunCommand(ctx, []string{"bundle", "plan", "--release", release, "--namespace", namespace, "--save-plan", planPath}, werf.CommonOptions{})

			By("applying the plan")
			dismissOnCleanup(werfProject, release, namespace)
			applyProject.RunCommand(ctx, []string{"bundle", "apply", "--use-plan", planPath}, werf.CommonOptions{})

			By("checking the deployed resources")
			cm, err := clientFactory.Static().CoreV1().ConfigMaps(namespace).Get(ctx, "app-cfg", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cm.Annotations).To(HaveKeyWithValue("project.werf.io/name", SuiteData.ProjectName))
		},
	)

	It("should deploy a published bundle as a dependency of a parent chart",
		func(ctx SpecContext) {
			clientFactory := werf.NewKubeClientFactory(ctx)
			dependencyRepo := suite_init.TestRepo(SuiteData.ProjectName) + "/mydep"

			By("publishing the dependency")
			werfProject.BundlePublish(ctx, &werf.BundlePublishOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: append([]string{"--repo", dependencyRepo, "--tag", "1.0.0"}, publishInputsArgs...),
				},
			})

			By("adding the pulled dependency to the parent chart")
			SuiteData.InitTestRepo(ctx, "parent", "werf-data/parent")
			parentProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("parent"))
			Expect(os.MkdirAll(filepath.Join(parentProject.GitRepoPath, ".helm/charts"), os.ModePerm)).To(Succeed())
			parentProject.RunCommand(ctx, []string{"helm", "pull", "--plain-http", "oci://" + dependencyRepo, "--version", "1.0.0", "--destination", ".helm/charts"}, werf.CommonOptions{})
			utils.RunSucceedCommand(ctx, parentProject.GitRepoPath, "git", "add", ".")
			utils.RunSucceedCommand(ctx, parentProject.GitRepoPath, "git", "commit", "-m", "add dependency")

			By("converging the parent chart")
			dismissOnCleanup(parentProject, parentProject.Release(ctx), parentProject.Namespace(ctx))
			parentProject.Converge(ctx, nil)

			By("checking the dependency's resources")
			cm, err := clientFactory.Static().CoreV1().ConfigMaps(parentProject.Namespace(ctx)).Get(ctx, "app-cfg", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cm.Data).To(Equal(map[string]string{
				"fromValues":       "from-values-file",
				"fromSecretValues": "from-secret-values-file",
				"fromSet":          "from-set",
				"fromSetFile":      "from-set-file",
				"greeting":         "parent-override",
				"werfEnv":          "test",
			}))

			By("checking that the dependency gets global.werf.* from the parent chart and keeps its own werf.*")
			werfValuesCM, err := clientFactory.Static().CoreV1().ConfigMaps(parentProject.Namespace(ctx)).Get(ctx, "app-werf-values", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(werfValuesCM.Data).To(Equal(map[string]string{
				"werfCommitHash":       strings.TrimSpace(utils.SucceedCommandOutputString(ctx, werfProject.GitRepoPath, "git", "rev-parse", "HEAD")),
				"globalWerfCommitHash": strings.TrimSpace(utils.SucceedCommandOutputString(ctx, parentProject.GitRepoPath, "git", "rev-parse", "HEAD")),
			}))
		},
	)

	It("should upgrade a release of a bundle published by werf 3.12 and roll it back",
		func(ctx SpecContext) {
			clientFactory := werf.NewKubeClientFactory(ctx)
			release, namespace := werfProject.Release(ctx), werfProject.Namespace(ctx)
			bundleRepo := suite_init.TestRepo(SuiteData.ProjectName) + "/v1app"
			setSecretKeyEnv(werfProject)

			getConfigMap := func() *corev1.ConfigMap {
				GinkgoHelper()

				cm, err := clientFactory.Static().CoreV1().ConfigMaps(namespace).Get(ctx, "app-cfg", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				return cm
			}

			By("pushing the bundle in the werf 3.12 format")
			chartsDir := SuiteData.GetTestRepoPath("v1-bundle")
			utils.CopyIn(utils.FixturePath("werf-data/v1-bundle"), chartsDir)
			applyProject.RunCommand(ctx, []string{"helm", "package", filepath.Join(chartsDir, "v1app"), "--destination", chartsDir}, werf.CommonOptions{})
			chartArchive, err := os.ReadFile(filepath.Join(chartsDir, "v1app-1.0.0.tgz"))
			Expect(err).NotTo(HaveOccurred())
			registryClient, err := helmregistry.NewClient(helmregistry.ClientOptPlainHTTP())
			Expect(err).NotTo(HaveOccurred())
			_, err = registryClient.Push(chartArchive, bundleRepo+":1.0.0")
			Expect(err).NotTo(HaveOccurred())

			By("applying the bundle in the werf 3.12 format")
			dismissOnCleanup(werfProject, release, namespace)
			applyProject.BundleApply(ctx, release, namespace, &werf.BundleApplyOptions{
				CommonOptions: werf.CommonOptions{ExtraArgs: []string{"--repo", bundleRepo, "--tag", "1.0.0"}},
			})
			v1Release := getRelease(ctx, applyProject, release, namespace)
			Expect(v1Release.Release.Annotations).To(HaveKeyWithValue("project.werf.io/name", "werf-test-e2e-bundle-werf-data"))
			v1Values := v1Release.Values
			v1ConfigMapData := map[string]string{
				"fromValues":       "from-values-file",
				"fromSecretValues": "from-secret-values-file",
				"fromSet":          "from-set",
				"fromSetFile":      "from-set-file",
				"greeting":         "chart",
				"werfEnv":          "test",
			}
			cm := getConfigMap()
			Expect(cm.Data).To(Equal(v1ConfigMapData))
			Expect(cm.Annotations).To(HaveKeyWithValue("team", "v1"))
			Expect(cm.Annotations).To(HaveKeyWithValue("project.werf.io/name", "werf-test-e2e-bundle-werf-data"))
			Expect(cm.Labels).To(HaveKeyWithValue("tier", "v1"))

			By("publishing and applying the bundle with .werf-data over the same release")
			werfProject.BundlePublish(ctx, &werf.BundlePublishOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: append(
						append([]string{"--repo", bundleRepo, "--tag", "2.0.0", "--add-annotation", "team=v2", "--add-label", "tier=v2"}, publishInputsArgs...),
						"--set=fromSet=from-set-v2",
					),
				},
			})
			applyProject.BundleApply(ctx, release, namespace, &werf.BundleApplyOptions{
				CommonOptions: werf.CommonOptions{ExtraArgs: []string{"--repo", bundleRepo, "--tag", "2.0.0"}},
			})
			releaseGet := getRelease(ctx, applyProject, release, namespace)
			Expect(releaseGet.Release.Revision).To(Equal(2))
			Expect(releaseGet.Values).To(HaveKeyWithValue("fromSet", "from-set-v2"))
			cm = getConfigMap()
			Expect(cm.Data).To(Equal(map[string]string{
				"fromValues":       "from-values-file",
				"fromSecretValues": "from-secret-values-file",
				"fromSet":          "from-set-v2",
				"fromSetFile":      "from-set-file",
				"greeting":         "chart",
				"werfEnv":          "test",
			}))
			Expect(cm.Annotations).To(HaveKeyWithValue("team", "v2"))
			Expect(cm.Labels).To(HaveKeyWithValue("tier", "v2"))

			By("rolling back to the bundle in the werf 3.12 format")
			applyProject.RunCommand(ctx, []string{"rollback", "--release", release, "--namespace", namespace, "--revision", "1"}, werf.CommonOptions{})
			releaseGet = getRelease(ctx, applyProject, release, namespace)
			Expect(releaseGet.Release.Revision).To(Equal(3))
			Expect(releaseGet.Values).To(Equal(v1Values))
			cm = getConfigMap()
			Expect(cm.Data).To(Equal(v1ConfigMapData))
			Expect(cm.Annotations).To(HaveKeyWithValue("team", "v1"))
			Expect(cm.Labels).To(HaveKeyWithValue("tier", "v1"))

			By("applying the bundle in the werf 3.12 format with service annotations overriding the bundle's")
			applyProject.BundleApply(ctx, release, namespace, &werf.BundleApplyOptions{
				CommonOptions: werf.CommonOptions{ExtraArgs: []string{
					"--repo", bundleRepo, "--tag", "1.0.0", "--env", "other",
					"--release-info-annotations", "project.werf.io/name=other",
					"--runtime-annotations", "project.werf.io/name=other",
					"--add-annotation", "ci.werf.io/commit=from-cli",
				}},
			})
			releaseGet = getRelease(ctx, applyProject, release, namespace)
			Expect(releaseGet.Release.Revision).To(Equal(4))
			Expect(releaseGet.Release.Annotations).To(HaveKeyWithValue("project.werf.io/name", "other"))
			Expect(releaseGet.Release.Annotations).To(HaveKeyWithValue("project.werf.io/env", "other"))
			Expect(releaseGet.Release.Annotations).To(HaveKeyWithValue("ci.werf.io/commit", "from-cli"))
			cm = getConfigMap()
			Expect(cm.Annotations).To(HaveKeyWithValue("project.werf.io/name", "other"))
			Expect(cm.Annotations).To(HaveKeyWithValue("project.werf.io/env", "other"))
		},
	)
})

type werfDataReleaseGetOutput struct {
	Release struct {
		Annotations map[string]string `json:"annotations"`
		Revision    int               `json:"revision"`
	} `json:"release"`
	Values map[string]any `json:"values"`
}

func getRelease(ctx context.Context, project *werf.Project, release, namespace string) werfDataReleaseGetOutput {
	GinkgoHelper()

	var get werfDataReleaseGetOutput
	out := project.RunCommand(ctx, []string{"release", "get", "--release", release, "--namespace", namespace, "--print-values", "--output-format", "json"}, werf.CommonOptions{})
	Expect(json.Unmarshal([]byte(out), &get)).To(Succeed())

	return get
}

func dismissOnCleanup(project *werf.Project, release, namespace string) {
	DeferCleanup(func(ctx SpecContext) {
		project.DismissAndDeleteNamespace(ctx, &werf.DismissOptions{
			Release:   release,
			Namespace: namespace,
		})
	})
}

// werf bundle apply does not read the .werf_secret_key of the project it runs outside of.
func setSecretKeyEnv(project *werf.Project) {
	GinkgoHelper()

	key, err := os.ReadFile(filepath.Join(project.GitRepoPath, ".werf_secret_key"))
	Expect(err).NotTo(HaveOccurred())
	SuiteData.Stubs.SetEnv("WERF_SECRET_KEY", strings.TrimSpace(string(key)))
}
