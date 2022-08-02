package deploy_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

var _ = Describe("Three way merge patches applier", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	AfterEach(func() {
		utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
	})

	Context("when deploying a resource with explicitly set replicas field", func() {
		It("should reset replicas to the specified number in the templates in converge if replicas has been changed manually or by the HPA", func() {
			convergeManualChangeThenConverge("three_way_merge_patches_applier_app1-001", 3, 3)
		})
	})

	Context("when deploying a resource with the special werf annotation to set replicas only on first creation and without explicitly set spec.replicas", func() {
		It("should set replicas only on first creation to the specified number in the templates in converge, should not reset replicas which was changed manually of by a HPA", func() {
			convergeManualChangeThenConverge("three_way_merge_patches_applier_app1-002", 3, 4)
		})
	})

	Context("when deploying a resource with the special werf annotation to set replicas only on first creation and with explicitly set spec.replicas", func() {
		It("should reset replicas to the specified number in the template spec.replicas in converge if replicas has been changed manually or by the HPA", func() {
			convergeManualChangeThenConverge("three_way_merge_patches_applier_app1-003", 3, 2)
		})
	})
})

func convergeManualChangeThenConverge(worktreeFixtureDir string, expectedReplicasAfterFirstConverge, expectedReplicasAfterSecondConverge int32) {
	namespace := SuiteData.ProjectName

	SuiteData.CommitProjectWorktree(SuiteData.ProjectName, worktreeFixtureDir, "initial commit")

	By("Installing release first time")

	Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

	mydeploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(*mydeploy.Spec.Replicas).To(Equal(expectedReplicasAfterFirstConverge))

	By("Changing replicas field manually through api")

	*mydeploy.Spec.Replicas = 4
	_, err = kube.Client.AppsV1().Deployments(namespace).Update(context.Background(), mydeploy, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())

	mydeploy, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(*mydeploy.Spec.Replicas).To(Equal(int32(4)))

	By("Reinstalling release second time")

	Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

	mydeploy, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(*mydeploy.Spec.Replicas).To(Equal(expectedReplicasAfterSecondConverge))
}
