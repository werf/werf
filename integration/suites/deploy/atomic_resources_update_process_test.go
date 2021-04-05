package deploy_test

import (
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/integration/pkg/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Atomic resources update process", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	AfterEach(func() {
		utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
	})

	Context("when deploying a new version of the chart with the newly added resource and an error in some other resource, then deploying the previous version of the chart without errors", func() {
		It("should create new resource on the first deploy, fail first deploy with an error, delete resource on the second deploy and succeed on second deploy", func() {
			namespace := SuiteData.ProjectName
			worktreeFixtureDir := "atomic_resources_update_process_app1-001"

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, worktreeFixtureDir, "initial commit")

			By("Installing release first time")

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

		})
	})

	Context("when deploying a new version of the chart with the newly added field of existing resource with some error in this field, then deploying the previous version of the chart without errors", func() {
		It("should fail on first deploy with an error, should revert newly added field change introduced into the resource on the second deploy and succeed on second deploy", func() {
		})
	})
})
