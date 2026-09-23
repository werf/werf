package e2e_converge_test

import (
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = ginkgo.Describe("Release lifecycle", ginkgo.Label("e2e", "converge", "simple", "release-lifecycle"), func() {
	var werfProject *werf.Project

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if werfProject == nil {
			return
		}

		werfProject.KubeCtl(ctx, &werf.KubeCtlOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{"delete", "namespace", "--ignore-not-found", werfProject.Namespace(ctx)},
			},
		})
	})

	ginkgo.It("should keep release revisions, rollback, plan and dismiss consistent",
		func(ctx ginkgo.SpecContext) {
			ginkgo.By("initializing")
			repoDirname := "repo0"
			setupEnv()

			clientFactory := werf.NewKubeClientFactory(ctx)

			ginkgo.By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "simple/state0")
			werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			listArgs := []string{"release", "list", "--namespace", werfProject.Namespace(ctx), "--output-format", "json", "--log-quiet"}
			releaseArgs := func(args ...string) []string {
				return append(args, "--release", werfProject.Release(ctx), "--namespace", werfProject.Namespace(ctx), "--output-format", "json", "--log-quiet")
			}

			ginkgo.By("state0: execute converge")
			werfProject.Converge(ctx, nil)

			ginkgo.By("state0: check deployed configmap")
			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value1"}))

			ginkgo.By("state1: preparing test repo")
			SuiteData.UpdateTestRepo(ctx, repoDirname, "simple/state1")

			ginkgo.By("state1: execute converge")
			werfProject.Converge(ctx, nil)

			ginkgo.By("state1: check deployed configmap")
			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value2"}))

			ginkgo.By("state1: check release list")
			var list releasesOutput
			gomega.Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, listArgs, werf.CommonOptions{})), &list)).To(gomega.Succeed())
			gomega.Expect(list.Releases).To(gomega.ContainElement(releaseSummary{
				Name:      werfProject.Release(ctx),
				Namespace: werfProject.Namespace(ctx),
				Revision:  2,
				Status:    "deployed",
			}))

			ginkgo.By("state1: check release history")
			gomega.Expect(releaseRevisions(ctx, werfProject, releaseArgs("release", "history"))).To(gomega.Equal([]releaseSummary{
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 1, Status: "superseded"},
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 2, Status: "deployed"},
			}))

			ginkgo.By("state1: check stored manifests of both revisions")
			gomega.Expect(releaseConfigMapData(ctx, werfProject, releaseArgs("release", "get", "1"))).To(gomega.Equal(map[string]string{"key1": "value1"}))
			gomega.Expect(releaseConfigMapData(ctx, werfProject, releaseArgs("release", "get", "2"))).To(gomega.Equal(map[string]string{"key1": "value2"}))

			ginkgo.By("rollback: execute rollback to revision 1")
			werfProject.RunCommand(ctx, []string{"rollback", "--revision", "1"}, werf.CommonOptions{})

			ginkgo.By("rollback: check configmap in cluster is rolled back")
			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value1"}))

			ginkgo.By("rollback: check new revision is recorded")
			gomega.Expect(releaseRevisions(ctx, werfProject, releaseArgs("release", "history"))).To(gomega.Equal([]releaseSummary{
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 1, Status: "superseded"},
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 2, Status: "superseded"},
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 3, Status: "deployed"},
			}))
			gomega.Expect(releaseConfigMapData(ctx, werfProject, releaseArgs("release", "get", "3"))).To(gomega.Equal(map[string]string{"key1": "value1"}))

			ginkgo.By("plan: check pending changes to the committed state1 are reported with exit code 2")
			planOut, planErr := utils.RunCommandWithOptions(ctx, SuiteData.GetTestRepoPath(repoDirname), SuiteData.WerfBinPath, []string{"plan", "--exit-code"}, utils.RunCommandOptions{})

			var planExitErr *exec.ExitError
			gomega.Expect(errors.As(planErr, &planExitErr)).To(gomega.BeTrue(), "expected werf plan to exit with a non-zero status, got: %v", planErr)
			gomega.Expect(planExitErr.ExitCode()).To(gomega.Equal(2))
			gomega.Expect(string(planOut)).To(gomega.And(gomega.ContainSubstring("value1"), gomega.ContainSubstring("value2")))

			ginkgo.By("plan: check cluster is left untouched")
			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value1"}))

			ginkgo.By("dismiss: execute dismiss")
			werfProject.DismissAndDeleteNamespace(ctx, nil)

			ginkgo.By("dismiss: check release is gone")
			var listAfterDismiss releasesOutput
			gomega.Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, listArgs, werf.CommonOptions{})), &listAfterDismiss)).To(gomega.Succeed())
			gomega.Expect(listAfterDismiss.Releases).To(gomega.BeEmpty())
		},
	)
})
