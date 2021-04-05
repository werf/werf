package deploy_test

import (
	"context"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/integration/pkg/utils/liveexec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Atomic resources update process", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	AfterEach(func() {
		utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
	})

	Context("when deploying a new version of the chart with the newly added resource and an error in some other resource, then deploying the previous version of the chart without errors", func() {
		It("should create new resource on the first deploy, fail first deploy with an error, delete resource on the second deploy and succeed on second deploy", func() {
			namespace := SuiteData.ProjectName

			By("Installing release first time")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-001", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			By("Upgrading release with an error and newly added resource in the chart")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-002", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).NotTo(Succeed())

			{
				stsList, err := kube.Client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundSts := false
				for _, item := range stsList.Items {
					if item.Name == "mysts" {
						foundSts = true
						break
					}
				}
				Expect(foundSts).To(BeTrue(), "expected StatefulSet mysts to exist")
			}

			By("Upgrading release without errors in the chart")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-001", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				stsList, err := kube.Client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundSts := false
				for _, item := range stsList.Items {
					if item.Name == "mysts" {
						foundSts = true
						break
					}
				}
				Expect(foundSts).To(BeFalse(), "expected StatefulSet mysts not to exist")
			}

		})
	})

	Context("when deploying a new version of the chart with the newly added field of existing resource with some error in this field, then deploying the previous version of the chart without errors", func() {
		It("should fail on first deploy with an error, should revert newly added field change introduced into the resource on the second deploy and succeed on second deploy", func() {
		})
	})
})
