package releaseserver_test

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/integration/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			utils.RunCommand("helm_releases_manager_app1-001", SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should keep no more than specified number of releases", func() {
			for i := 0; i < 9; i++ {
				Expect(werfDeploy("helm_releases_manager_app1-001", liveexec.ExecCommandOptions{
					Env: map[string]string{"WERF_RELEASES_HISTORY_MAX": "5"},
				})).Should(Succeed())
				Expect(len(getReleasesHistory(releaseName, releaseName)) <= 5).To(BeTrue())
			}
			Expect(len(getReleasesHistory(releaseName, releaseName))).To(Equal(5))
		})
	})

	Context("when releases-history-max was not specified initially and then specified", func() {
		AfterEach(func() {
			utils.RunCommand("helm_releases_manager_app1-001", SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should keep no more than specified number of releases", func() {
			for i := 0; i < 9; i++ {
				Expect(werfDeploy("helm_releases_manager_app1-001", liveexec.ExecCommandOptions{})).Should(Succeed())
			}
			Expect(len(getReleasesHistory(releaseName, releaseName))).To(Equal(9))

			for i := 0; i < 5; i++ {
				Expect(werfDeploy("helm_releases_manager_app1-001", liveexec.ExecCommandOptions{}, "--releases-history-max=5")).Should(Succeed())
				Expect(len(getReleasesHistory(releaseName, releaseName))).To(Equal(5))
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
