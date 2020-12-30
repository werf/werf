package bundles_test

import (
	"fmt"
	"os"

	"github.com/werf/werf/integration/suite_init"

	"github.com/werf/werf/integration/utils"

	"github.com/werf/werf/integration/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func liveExecWerf(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(extraArgs...)...)
}

var _ = Describe("Bundles", func() {
	Context("working with simple werf project", func() {
		It("should publish bundle then apply into kubernetes", func() {
			//TODO
			//Expect(liveExecWerf("bundles_app1-001", liveexec.ExecCommandOptions{}, "bundle", "publish", "--repo", SuiteData.K8sDockerRegistryRepo)).Should(Succeed())
		})
	})
})
