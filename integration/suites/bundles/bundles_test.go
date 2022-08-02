package bundles_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/suite_init"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

func liveExecWerf(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(extraArgs...)...)
}

var _ = Describe("Bundles", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(false) {
		implementationName := iName

		Context(fmt.Sprintf("[%s] publish and apply quickstart-application bundle", implementationName), func() {
			BeforeEach(func() {
				SuiteData.Repo = fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
				SuiteData.SetupRepo(context.Background(), SuiteData.Repo, implementationName, SuiteData.StubsData)
			})

			AfterEach(func() {
				liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "helm", "uninstall", "--namespace", SuiteData.ProjectName, SuiteData.ProjectName)

				kube.Client.CoreV1().Namespaces().Delete(context.Background(), SuiteData.ProjectName, metav1.DeleteOptions{})

				liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "host", "purge", "--force")
				os.RemoveAll(SuiteData.ProjectName)

				SuiteData.TeardownRepo(context.Background(), SuiteData.Repo, implementationName, SuiteData.StubsData)
			})

			It("should publish latest quickstart-application bundle, then apply into kubernetes, then export it, then render it from exported dir, then render it from registry", func() {
				ctx := context.Background()

				switch implementationName {
				case "dockerhub":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3184")
				case "quay":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3182")
				}

				const secretKey = "bfd966688bbe64c1986a356be2d6ba0a"

				By("preparing test application using werf/quickstart-application as a base")

				Expect(liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, []string{"clone", "https://github.com/werf/quickstart-application", SuiteData.ProjectName}...)).To(Succeed())

				Expect(os.WriteFile(filepath.Join(SuiteData.ProjectName, ".helm", "templates", "secret.yaml"), []byte(`apiVersion: v1
kind: Secret
metadata:
  name: test-secret
data:
  testkey: {{ .Values.testsecrets.testkey | b64enc }}
`), os.ModePerm)).To(Succeed())

				Expect(os.WriteFile(filepath.Join(SuiteData.ProjectName, ".helm", "values.yaml"), []byte(`testsecrets:
  testkey: DEFAULT
`), os.ModePerm)).To(Succeed())

				Expect(liveexec.ExecCommand(SuiteData.ProjectName, "git", liveexec.ExecCommandOptions{}, []string{"add", "."}...)).To(Succeed())
				Expect(liveexec.ExecCommand(SuiteData.ProjectName, "git", liveexec.ExecCommandOptions{}, []string{"commit", "-m", "go"}...)).To(Succeed())

				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "bundle", "publish")).Should(Succeed())

				By("applying bundle into cluster")

				applyDir := filepath.Join(SuiteData.TmpDir, "apply-dir")
				Expect(os.MkdirAll(applyDir, os.ModePerm)).To(Succeed())
				secretFile := filepath.Join(applyDir, "secret-values.yaml")
				Expect(os.WriteFile(secretFile, []byte(`
testsecrets:
  testkey: 1000b45ee4272d14b30be2d20b5963f09e372fdfe761bf3913186938f4054d09ed0e
`), os.ModePerm)).To((Succeed()))

				Expect(liveExecWerf(".", liveexec.ExecCommandOptions{
					Env: map[string]string{"WERF_SECRET_KEY": secretKey},
				}, "bundle", "apply", "--release", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName, "--set-docker-config-json-value", "--secret-values", secretFile)).Should(Succeed())

				{
					secret, err := kube.Client.CoreV1().Secrets(SuiteData.ProjectName).Get(ctx, "test-secret", metav1.GetOptions{})
					Expect(err).To(Succeed())
					Expect(string(secret.Data["testkey"])).To(Equal("TOPSECRET"))
				}

				By("exporting bundle")

				exportedBundleDir := utils.GetTempDir()
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "bundle", "export", "--destination", exportedBundleDir)).Should(Succeed())

				SuiteData.Stubs.UnsetEnv("WERF_REPO")

				By("rendering bundle from the directory which contains exported bundle")

				gotKindDeploymentInLocalRender := false
				Expect(liveExecWerf(exportedBundleDir, liveexec.ExecCommandOptions{
					Env: map[string]string{"WERF_SECRET_KEY": secretKey},
					OutputLineHandler: func(line string) {
						if strings.TrimSpace(line) == "kind: Deployment" {
							gotKindDeploymentInLocalRender = true
						}
					},
				}, "bundle", "render", "--bundle-dir", ".")).Should(Succeed())
				Expect(gotKindDeploymentInLocalRender).To(BeTrue())

				SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.Repo)

				By("rendering bundle from the project dir")

				gotKindDeploymentInRemoteRender := false
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{
					Env: map[string]string{"WERF_SECRET_KEY": secretKey},
					OutputLineHandler: func(line string) {
						if strings.TrimSpace(line) == "kind: Deployment" {
							gotKindDeploymentInRemoteRender = true
						}
					},
				}, "bundle", "render")).Should(Succeed())
				Expect(gotKindDeploymentInRemoteRender).To(BeTrue())
			})
		})
	}
})
