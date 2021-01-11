package bundles_test

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/docker_registry"

	"github.com/werf/werf/integration/pkg/suite_init"

	"github.com/werf/werf/integration/pkg/utils"

	"github.com/werf/werf/integration/pkg/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func liveExecWerf(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(extraArgs...)...)
}

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))

var _ = XDescribe("Bundles", func() {
	SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
		implementations := suite_init.ContainerRegistryImplementationListToCheck()
		Expect(len(implementations)).NotTo(Equal(0), "expected at least one of WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_<IMPLEMENTATION>=1 to be set, supported implementations: %v", docker_registry.ImplementationList())
	})

	for _, implementationName := range suite_init.ContainerRegistryImplementationListToCheck() {
		Context(fmt.Sprintf("[%s] publish and apply quickstart-application bundle", implementationName), func() {
			SuiteData.SetupContainerRegistryPerImplementation(suite_init.NewContainerRegistryPerImplementationData(SuiteData.ProjectNameData, SuiteData.StubsData, implementationName))

			AfterEach(func() {
				liveExecWerf("quickstart-application", liveexec.ExecCommandOptions{}, "purge", "--force")
				os.RemoveAll("quickstart-application")
			})

			It("should publish latest quickstart-application bundle then apply into kubernetes", func() {
				Expect(liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, append([]string{"clone", "https://github.com/werf/quickstart-application", fmt.Sprintf("quickstart-application-%s", SuiteData.ProjectName)})...)).To(Succeed())
				Expect(liveExecWerf(fmt.Sprintf("quickstart-application-%s", SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "bundle", "publish", "--repo", SuiteData.ContainerRegistryPerImplementation[implementationName].Repo)).Should(Succeed())
				Expect(liveExecWerf(".", liveexec.ExecCommandOptions{}, "bundle", "apply", "--repo", SuiteData.ContainerRegistryPerImplementation[implementationName].Repo, "--release", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName)).Should(Succeed())
			})
		})
	}
})
