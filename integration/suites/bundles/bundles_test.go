package bundles_test

import (
	"fmt"
	"os"

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
	for i := range suite_init.ContainerRegistryImplementationListToCheck() {
		implementationName := suite_init.ContainerRegistryImplementationListToCheck()[i]

		Context(fmt.Sprintf("[%s] publish and apply quickstart-application bundle", implementationName), func() {
			AfterEach(func() {
				liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "purge", "--force")
				os.RemoveAll(SuiteData.ProjectName)
			})

			It("should publish latest quickstart-application bundle then apply into kubernetes", func() {
				SuiteData.ContainerRegistryPerImplementationData.ActivateImplementationWerfEnvironmentParams(implementationName, SuiteData.StubsData)

				repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)

				Expect(liveexec.ExecCommand(".", "git", liveexec.ExecCommandOptions{}, append([]string{"clone", "https://github.com/werf/quickstart-application", SuiteData.ProjectName})...)).To(Succeed())
				Expect(liveExecWerf(SuiteData.ProjectName, liveexec.ExecCommandOptions{}, "bundle", "publish", "--repo", repo)).Should(Succeed())
				Expect(liveExecWerf(".", liveexec.ExecCommandOptions{}, "bundle", "apply", "--repo", repo, "--release", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName)).Should(Succeed())
			})
		})
	}
})
