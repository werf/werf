package bundles_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
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
				switch implementationName {
				case "dockerhub":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3184")
				case "quay":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3182")
				}

				Expect(liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, []string{"clone", "https://github.com/werf/quickstart-application", SuiteData.ProjectName}...)).To(Succeed())
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "bundle", "publish")).Should(Succeed())
				Expect(liveExecWerf(".", liveexec.ExecCommandOptions{}, "bundle", "apply", "--release", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName, "--set-docker-config-json-value")).Should(Succeed())

				exportedBundleDir := utils.GetTempDir()
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "bundle", "export", "--destination", exportedBundleDir)).Should(Succeed())

				SuiteData.Stubs.UnsetEnv("WERF_REPO")

				gotKindDeploymentInLocalRender := false
				Expect(liveExecWerf(exportedBundleDir, liveexec.ExecCommandOptions{
					OutputLineHandler: func(line string) {
						if strings.TrimSpace(line) == "kind: Deployment" {
							gotKindDeploymentInLocalRender = true
						}
					},
				}, "bundle", "render", "--bundle-dir", ".")).Should(Succeed())
				Expect(gotKindDeploymentInLocalRender).To(BeTrue())

				SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.Repo)

				gotKindDeploymentInRemoteRender := false
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{
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
