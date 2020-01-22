package releaseserver_test

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helm releases manager", func() {
	var projectName, releaseName string

	BeforeEach(func() {
		projectName = utils.ProjectName()
		releaseName = fmt.Sprintf("%s-dev", projectName)
	})

	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when releases-history-max option has been specified from the beginning", func() {
		AfterEach(func() {
			utils.RunCommand("helm_releases_manager_app1-001", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should keep no more than specified number of releases", func() {
			for i := 0; i < 20; i++ {
				Expect(werfDeploy("helm_releases_manager_app1-001", liveexec.ExecCommandOptions{
					Env: map[string]string{"WERF_RELEASES_HISTORY_MAX": "5"},
				})).Should(Succeed())
				Expect(len(getReleasesHistory(releaseName)) <= 5).To(BeTrue())
			}
			Expect(len(getReleasesHistory(releaseName))).To(Equal(5))
		})
	})

	Context("when releases-history-max was not specified initially and then specified", func() {
		AfterEach(func() {
			utils.RunCommand("helm_releases_manager_app1-001", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should keep no more than specified number of releases", func() {
			for i := 0; i < 20; i++ {
				Expect(werfDeploy("helm_releases_manager_app1-001", liveexec.ExecCommandOptions{})).Should(Succeed())
			}
			Expect(len(getReleasesHistory(releaseName))).To(Equal(20))

			for i := 0; i < 5; i++ {
				Expect(werfDeploy("helm_releases_manager_app1-001", liveexec.ExecCommandOptions{}, "--releases-history-max=5")).Should(Succeed())
				Expect(len(getReleasesHistory(releaseName))).To(Equal(5))
			}
		})
	})
})

func getReleasesHistory(releaseName string) []*corev1.ConfigMap {
	cmList, err := kube.Kubernetes.CoreV1().ConfigMaps("kube-system").List(metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	var releases []*corev1.ConfigMap

	for i := range cmList.Items {
		item := cmList.Items[i]
		if strings.HasPrefix(item.Name, fmt.Sprintf("%s.v", releaseName)) {
			releases = append(releases, &item)
			_, _ = fmt.Fprintf(GinkgoWriter, "[DEBUG] RELEASE LISTING ITEM: cm/%s\n", item.Name)
		}
	}

	return releases
}
