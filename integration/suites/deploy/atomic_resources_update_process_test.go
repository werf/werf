package deploy_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
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

			By("Installing release first time")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-001", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			By("Upgrading release with an error and newly added resource in the chart")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-003", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).NotTo(Succeed())

			{
				deployList, err := kube.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundDeploy := false
				for _, item := range deployList.Items {
					if item.Name == "mydeploy" {
						foundDeploy = true
						break
					}
				}
				Expect(foundDeploy).To(BeTrue(), "expected Deployment mydeploy to exist")
			}

			By("Upgrading release without errors in the chart")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-001", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				deployList, err := kube.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundDeploy := false
				for _, item := range deployList.Items {
					if item.Name == "mydeploy" {
						foundDeploy = true
						break
					}
				}
				Expect(foundDeploy).To(BeFalse(), "expected Deployment mydeploy not to exist")
			}
		})
	})

	Context("when deploying a new version of the chart with the newly added field of existing resource with some error in this field, then deploying the previous version of the chart without errors", func() {
		It("should fail on first deploy with an error, should revert newly added field change introduced into the resource on the second deploy and succeed on second deploy", func() {
			namespace := SuiteData.ProjectName

			By("Installing release first time")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-001", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			By("Upgrading release with newly added resource in the chart")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-002", "add good resource")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				deployList, err := kube.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundDeploy := false
				for _, item := range deployList.Items {
					if item.Name == "mydeploy" {
						foundDeploy = true
						break
					}
				}
				Expect(foundDeploy).To(BeTrue(), "expected Deployment mydeploy to exist")
			}

			{
				deploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(strings.Join(deploy.Spec.Template.Spec.Containers[0].Command, " ")).To(Equal("/bin/sh -ec while true ; do date ; sleep 1 ; done"))
			}

			By("Upgrade release, introduce error into the previously added resource")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-003", "change resource to bad")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).NotTo(Succeed())

			{
				deployList, err := kube.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundDeploy := false
				for _, item := range deployList.Items {
					if item.Name == "mydeploy" {
						foundDeploy = true
						break
					}
				}
				Expect(foundDeploy).To(BeTrue(), "expected Deployment mydeploy to exist")
			}

			{
				deploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(len(deploy.Spec.Template.Spec.InitContainers)).To(Equal(1))
			}

			By("Upgrade release with reverted breaking change, which was introduced in the previous step")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "atomic_resources_update_process_app1-002", "fix bad resource")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				deployList, err := kube.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				foundDeploy := false
				for _, item := range deployList.Items {
					if item.Name == "mydeploy" {
						foundDeploy = true
						break
					}
				}
				Expect(foundDeploy).To(BeTrue(), "expected Deployment mydeploy to exist")
			}

			{
				deploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(len(deploy.Spec.Template.Spec.InitContainers)).To(Equal(0))
			}
		})
	})

	// TODO: initial release installation with an error, helm purge should not be needed
})
