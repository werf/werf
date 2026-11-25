package e2e_converge_test

import (
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/utils"
	"github.com/werf/werf/v3/test/pkg/werf"
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

			listArgs := []string{"release", "list", "--namespace", werfProject.Namespace(ctx), "--output-format", "json"}
			releaseArgs := func(args ...string) []string {
				return append(args, "--release", werfProject.Release(ctx), "--namespace", werfProject.Namespace(ctx), "--output-format", "json")
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

	ginkgo.It("should not write empty project annotations on rollback",
		func(ctx ginkgo.SpecContext) {
			ginkgo.By("initializing")
			repoDirname := "repo1"
			setupEnv()

			clientFactory := werf.NewKubeClientFactory(ctx)

			SuiteData.InitTestRepo(ctx, repoDirname, "simple/state0")
			werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			outsideDir := ginkgo.GinkgoT().TempDir()
			releaseFlags := []string{"--release", werfProject.Release(ctx), "--namespace", werfProject.Namespace(ctx)}
			rollbackArgs := append([]string{"rollback"}, releaseFlags...)
			releaseGetArgs := append([]string{"release", "get", "--output-format", "json"}, releaseFlags...)

			ginkgo.By("converge: deploy two revisions")
			werfProject.Converge(ctx, &werf.ConvergeOptions{})
			SuiteData.UpdateTestRepo(ctx, repoDirname, "simple/state1")
			werfProject.Converge(ctx, &werf.ConvergeOptions{})

			gomega.Expect(clusterConfigMapAnnotations(ctx, clientFactory, werfProject)).To(gomega.And(
				gomega.HaveKeyWithValue("project.werf.io/name", SuiteData.ProjectName),
				gomega.HaveKeyWithValue("project.werf.io/env", "test"),
			))

			ginkgo.By("rollback inside the project with --release and --namespace: check project annotations are kept")
			utils.RunSucceedCommand(ctx, werfProject.GitRepoPath, SuiteData.WerfBinPath, rollbackArgs...)

			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value1"}))
			gomega.Expect(clusterConfigMapAnnotations(ctx, clientFactory, werfProject)).To(gomega.And(
				gomega.HaveKeyWithValue("project.werf.io/name", SuiteData.ProjectName),
				gomega.HaveKeyWithValue("project.werf.io/env", "test"),
			), "configmap")
			gomega.Expect(releaseAnnotations(utils.SucceedCommandOutputString(ctx, werfProject.GitRepoPath, SuiteData.WerfBinPath, releaseGetArgs...))).To(gomega.And(
				gomega.HaveKeyWithValue("project.werf.io/name", SuiteData.ProjectName),
				gomega.HaveKeyWithValue("project.werf.io/env", "test"),
			), "release")

			ginkgo.By("converge: deploy state1 before the rollback outside the project with --env")
			werfProject.Converge(ctx, &werf.ConvergeOptions{})

			ginkgo.By("rollback outside the project with --env: check only project.werf.io/env is written")
			utils.RunSucceedCommand(ctx, outsideDir, SuiteData.WerfBinPath, rollbackArgs...)

			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value1"}))
			gomega.Expect(clusterConfigMapAnnotations(ctx, clientFactory, werfProject)).To(gomega.And(
				gomega.Not(gomega.HaveKey("project.werf.io/name")),
				gomega.HaveKeyWithValue("project.werf.io/env", "test"),
			), "configmap")
			gomega.Expect(releaseAnnotations(utils.SucceedCommandOutputString(ctx, outsideDir, SuiteData.WerfBinPath, releaseGetArgs...))).To(gomega.And(
				gomega.Not(gomega.HaveKey("project.werf.io/name")),
				gomega.HaveKeyWithValue("project.werf.io/env", "test"),
			), "release")

			ginkgo.By("converge: deploy state1 before the rollback outside the project without --env")
			werfProject.Converge(ctx, &werf.ConvergeOptions{})

			ginkgo.By("rollback outside the project without --env: check no project annotations are written")
			SuiteData.Stubs.UnsetEnv("WERF_ENV")
			utils.RunSucceedCommand(ctx, outsideDir, SuiteData.WerfBinPath, rollbackArgs...)

			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value1"}))
			gomega.Expect(clusterConfigMapAnnotations(ctx, clientFactory, werfProject)).NotTo(gomega.Or(
				gomega.HaveKey("project.werf.io/name"),
				gomega.HaveKey("project.werf.io/env"),
			), "configmap")
			gomega.Expect(releaseAnnotations(utils.SucceedCommandOutputString(ctx, outsideDir, SuiteData.WerfBinPath, releaseGetArgs...))).NotTo(gomega.Or(
				gomega.HaveKey("project.werf.io/name"),
				gomega.HaveKey("project.werf.io/env"),
			), "release")

			ginkgo.By("rollback with --release, --namespace and a missing --config: check it fails without a new revision")
			_, err := utils.RunCommand(ctx, werfProject.GitRepoPath, SuiteData.WerfBinPath, append(rollbackArgs, "--config", "missing.yaml")...)
			gomega.Expect(err).To(gomega.HaveOccurred())

			var get releaseGetOutput
			gomega.Expect(json.Unmarshal([]byte(utils.SucceedCommandOutputString(ctx, outsideDir, SuiteData.WerfBinPath, releaseGetArgs...)), &get)).To(gomega.Succeed())
			gomega.Expect(get.Release.Revision).To(gomega.Equal(7))

			ginkgo.By("rollback with --release and --namespace from a git repo without werf.yaml: check it succeeds")
			gitDir := ginkgo.GinkgoT().TempDir()
			utils.RunSucceedCommand(ctx, gitDir, "git", "init", "-q")
			utils.RunSucceedCommand(ctx, gitDir, "git", "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "--allow-empty", "-m", "init")
			utils.RunSucceedCommand(ctx, gitDir, SuiteData.WerfBinPath, rollbackArgs...)

			gomega.Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(gomega.Equal(map[string]string{"key1": "value2"}))
		},
	)
})
