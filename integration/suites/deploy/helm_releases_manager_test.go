package deploy_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

var _ = Describe("Helm releases manager", func() {
	var projectName, releaseName string

	BeforeEach(func() {
		projectName = utils.ProjectName()
		releaseName = projectName
	})

	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when releases-history-max option has been specified from the beginning", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should keep no more than specified number of releases", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "helm_releases_manager_app1-001", "initial commit")

			for i := 0; i < 9; i++ {
				Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
					Env: map[string]string{"WERF_RELEASES_HISTORY_MAX": "5"},
				})).Should(Succeed())
				Expect(len(getReleasesHistory(releaseName, releaseName)) <= 5).To(BeTrue())
			}
			Expect(len(getReleasesHistory(releaseName, releaseName))).To(Equal(5))
		})
	})

	Context("when releases-history-max was not specified initially and then specified", func() {
		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should keep no more than specified number of releases", func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "helm_releases_manager_app1-001", "initial commit")

			for i := 0; i < 4; i++ {
				Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).Should(Succeed())
			}
			Expect(len(getReleasesHistory(releaseName, releaseName))).To(Equal(4))

			for i := 0; i < 2; i++ {
				Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--releases-history-max=2")).Should(Succeed())
				Expect(len(getReleasesHistory(releaseName, releaseName))).To(Equal(2))
			}
		})
	})
})

func getReleasesHistory(namespace, releaseName string) []*corev1.Secret {
	resourceList, err := kube.Kubernetes.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	var releases []*corev1.Secret

	for i := range resourceList.Items {
		item := resourceList.Items[i]

		if strings.HasPrefix(item.Name, fmt.Sprintf("sh.helm.release.v1.%s.v", releaseName)) {
			releases = append(releases, &item)
			_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG] RELEASE LISTING ITEM: cm/%s\n", item.Name)
		}
	}

	return releases
}
