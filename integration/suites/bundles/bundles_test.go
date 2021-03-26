package bundles_test

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"

	"github.com/werf/werf/integration/pkg/suite_init"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/integration/pkg/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

			It("should publish latest quickstart-application bundle then apply into kubernetes", func() {
				switch implementationName {
				case "dockerhub":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3184")
				case "github":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3188")
				case "quay":
					Skip("Skip due to the unresolved issue: https://github.com/werf/werf/issues/3182")
				}

				Expect(liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, append([]string{"clone", "https://github.com/werf/quickstart-application", SuiteData.ProjectName})...)).To(Succeed())
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "bundle", "publish")).Should(Succeed())
				Expect(liveExecWerf(".", liveexec.ExecCommandOptions{}, "bundle", "apply", "--release", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName, "--set-docker-config-json-value")).Should(Succeed())
			})
		})
	}
})
