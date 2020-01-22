package releaseserver_test

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources owner checker", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when three-way-merge is disabled, current release is in FAILED state and does not have owner-release references", func() {
		var namespace, projectName, releaseName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = fmt.Sprintf("%s-dev", projectName)
			releaseName = fmt.Sprintf("%s-dev", projectName)
		})

		AfterEach(func() {
			utils.RunCommand("resources_owner_checker_app1-003", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should set owner-release refs during rollback operation https://github.com/flant/werf/issues/1902", func() {
			By("creating deployed release in FAILED state without service.werf.io/owner-release annotations (emulating already existing old werf release)")

			Expect(werfDeploy("resources_owner_checker_app1-001", liveexec.ExecCommandOptions{}, "--three-way-merge-mode", "disabled")).To(Succeed())
			Expect(werfDeploy("resources_owner_checker_app1-002", liveexec.ExecCommandOptions{}, "--three-way-merge-mode", "disabled")).NotTo(Succeed())

		GetAndUpdateReleaseCm:
			releaseCm, err := kube.Kubernetes.CoreV1().ConfigMaps("kube-system").Get(fmt.Sprintf("%s.v2", releaseName), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			delete(releaseCm.Annotations, "werf.io/resources-has-owner-release-name")
			releaseCm, err = kube.Kubernetes.CoreV1().ConfigMaps("kube-system").Update(releaseCm)
			if errors.IsConflict(err) {
				goto GetAndUpdateReleaseCm
			}
			Expect(err).NotTo(HaveOccurred())

		GetAndUpdateMydeploy1:
			mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			delete(mydeploy1.Annotations, "service.werf.io/owner-release")
			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy1)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy1
			}
			Expect(err).NotTo(HaveOccurred())

			By("updating old release should set owner-release references to the existing resources")

			// Should succeed without "inconsistent state detected" error
			Expect(werfDeploy("resources_owner_checker_app1-003", liveexec.ExecCommandOptions{}, "--three-way-merge-mode", "disabled")).To(Succeed())

			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy1.Annotations["service.werf.io/owner-release"]).To(Equal(releaseName))
		})
	})
})
