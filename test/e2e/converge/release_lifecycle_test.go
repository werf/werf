package e2e_converge_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/nelm/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type releaseSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
}

type releasesOutput struct {
	Releases []releaseSummary `json:"releases"`
}

type releaseGetOutput struct {
	Release       releaseSummary `json:"release"`
	ResourceSpecs []struct {
		Unstruct struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Data map[string]string `json:"data"`
		} `json:"unstruct"`
	} `json:"resourceSpecs"`
}

var _ = Describe("Release lifecycle", Label("e2e", "converge", "simple", "release-lifecycle"), func() {
	var werfProject *werf.Project

	AfterEach(func(ctx SpecContext) {
		werfProject.KubeCtl(ctx, &werf.KubeCtlOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{"delete", "namespace", "--ignore-not-found", werfProject.Namespace(ctx)},
			},
		})
	})

	It("should keep release revisions, rollback, plan and dismiss consistent",
		func(ctx SpecContext) {
			By("initializing")
			repoDirname := "repo0"
			setupEnv()

			clientFactory := werf.NewKubeClientFactory(ctx)

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "simple/state0")
			werfProject = werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			listArgs := []string{"release", "list", "--namespace", werfProject.Namespace(ctx), "--output-format", "json", "--log-quiet"}
			releaseArgs := func(args ...string) []string {
				return append(args, "--release", werfProject.Release(ctx), "--namespace", werfProject.Namespace(ctx), "--output-format", "json", "--log-quiet")
			}

			By("state0: execute converge")
			werfProject.Converge(ctx, nil)

			By("state0: check deployed configmap")
			Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(Equal(map[string]string{"key1": "value1"}))

			By("state1: preparing test repo")
			SuiteData.UpdateTestRepo(ctx, repoDirname, "simple/state1")

			By("state1: execute converge")
			werfProject.Converge(ctx, nil)

			By("state1: check deployed configmap")
			Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(Equal(map[string]string{"key1": "value2"}))

			By("state1: check release list")
			var list releasesOutput
			Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, listArgs, werf.CommonOptions{})), &list)).To(Succeed())
			Expect(list.Releases).To(ContainElement(releaseSummary{
				Name:      werfProject.Release(ctx),
				Namespace: werfProject.Namespace(ctx),
				Revision:  2,
				Status:    "deployed",
			}))

			By("state1: check release history")
			Expect(releaseRevisions(ctx, werfProject, releaseArgs("release", "history"))).To(Equal([]releaseSummary{
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 1, Status: "superseded"},
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 2, Status: "deployed"},
			}))

			By("state1: check stored manifests of both revisions")
			Expect(releaseConfigMapData(ctx, werfProject, releaseArgs("release", "get", "1"))).To(Equal(map[string]string{"key1": "value1"}))
			Expect(releaseConfigMapData(ctx, werfProject, releaseArgs("release", "get", "2"))).To(Equal(map[string]string{"key1": "value2"}))

			By("rollback: execute rollback to revision 1")
			werfProject.RunCommand(ctx, []string{"rollback", "--revision", "1"}, werf.CommonOptions{})

			By("rollback: check configmap in cluster is rolled back")
			Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(Equal(map[string]string{"key1": "value1"}))

			By("rollback: check new revision is recorded")
			Expect(releaseRevisions(ctx, werfProject, releaseArgs("release", "history"))).To(Equal([]releaseSummary{
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 1, Status: "superseded"},
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 2, Status: "superseded"},
				{Name: werfProject.Release(ctx), Namespace: werfProject.Namespace(ctx), Revision: 3, Status: "deployed"},
			}))
			Expect(releaseConfigMapData(ctx, werfProject, releaseArgs("release", "get", "3"))).To(Equal(map[string]string{"key1": "value1"}))

			By("plan: check pending changes to the committed state1 are reported")
			planOut := werfProject.RunCommand(ctx, []string{"plan", "--exit-code"}, werf.CommonOptions{ShouldFail: true})
			Expect(planOut).To(ContainSubstring("value2"))

			By("plan: check cluster is left untouched")
			Expect(clusterConfigMapData(ctx, clientFactory, werfProject)).To(Equal(map[string]string{"key1": "value1"}))

			By("dismiss: execute dismiss")
			werfProject.DismissAndDeleteNamespace(ctx, nil)

			By("dismiss: check release is gone")
			var listAfterDismiss releasesOutput
			Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, listArgs, werf.CommonOptions{})), &listAfterDismiss)).To(Succeed())
			Expect(listAfterDismiss.Releases).To(BeEmpty())
		},
	)
})

func clusterConfigMapData(ctx SpecContext, clientFactory *kube.ClientFactory, werfProject *werf.Project) map[string]string {
	GinkgoHelper()

	cm, err := clientFactory.Static().CoreV1().ConfigMaps(werfProject.Namespace(ctx)).Get(ctx, "test1", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	return cm.Data
}

func releaseRevisions(ctx SpecContext, werfProject *werf.Project, args []string) []releaseSummary {
	GinkgoHelper()

	var history releasesOutput
	Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, args, werf.CommonOptions{})), &history)).To(Succeed())

	return history.Releases
}

func releaseConfigMapData(ctx SpecContext, werfProject *werf.Project, args []string) map[string]string {
	GinkgoHelper()

	var get releaseGetOutput
	Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, args, werf.CommonOptions{})), &get)).To(Succeed())

	for _, spec := range get.ResourceSpecs {
		if spec.Unstruct.Kind == "ConfigMap" && spec.Unstruct.Metadata.Name == "test1" {
			return spec.Unstruct.Data
		}
	}

	Fail("configmap test1 not found in release manifests")

	return nil
}
