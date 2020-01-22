package releaseserver_test

import (
	"fmt"
	"strings"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helm hooks deleter", func() {
	Context("when installing chart with post-install Job hook and hook-succeeded delete policy", func() {
		AfterEach(func() {
			utils.RunCommand("helm_hooks_deleter_app1", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should delete hook when hook succeeded and wait till it is deleted without timeout https://github.com/flant/werf/issues/1885", func() {
			gotDeletingHookLine := false

			Expect(werfDeploy("helm_hooks_deleter_app1", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.HasPrefix(line, "│ NOTICE Will not delete Job/migrate: resource does not belong to the helm release")).ShouldNot(BeTrue(), fmt.Sprintf("Got unexpected output line: %v", line))

					if strings.HasPrefix(line, "│ Deleting resource Job/migrate from release") {
						gotDeletingHookLine = true
					}
				},
			})).Should(Succeed())

			Expect(gotDeletingHookLine).Should(BeTrue())
		})
	})

	Context("when releasing a chart containing a hook with before-hook-creation delete policy", func() {
		var namespace, projectName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = fmt.Sprintf("%s-dev", projectName)
		})

		BeforeEach(func() {
			Expect(kube.Init(kube.InitOptions{})).To(Succeed())
		})

		AfterEach(func() {
			utils.RunCommand("helm_hooks_deleter_app2", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should create hook on release install, delete hook on next release upgrade due to before-hook-creation delete policy", func() {
			hookName := "myhook"

			Expect(werfDeploy("helm_hooks_deleter_app2", liveexec.ExecCommandOptions{})).Should(Succeed())

		GetAndUpdate:
			hookObj, err := kube.Kubernetes.BatchV1().Jobs(namespace).Get(hookName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			delete(hookObj.Annotations, "service.werf.io/owner-release")
			newHookObj, err := kube.Kubernetes.BatchV1().Jobs(namespace).Update(hookObj)
			if errors.IsConflict(err) {
				goto GetAndUpdate
			}
			Expect(err).NotTo(HaveOccurred())

			Expect(newHookObj.Annotations["service.werf.io/owner-release"]).To(BeEmpty())
			hookObj = newHookObj

			gotDeletingHookLine := false
			// Update release, hook should be deleted by before-hook-creation policy and created again
			Expect(werfDeploy("helm_hooks_deleter_app2", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.HasPrefix(line, "│ NOTICE Will not delete Job/myhook: resource does not belong to the helm release")).ShouldNot(BeTrue(), fmt.Sprintf("Got unexpected output line: %v", line))

					if strings.HasPrefix(line, "│ Deleting resource Job/myhook from release") {
						gotDeletingHookLine = true
					}
				},
			})).Should(Succeed())
			Expect(gotDeletingHookLine).Should(BeTrue())

			newHookObj, err = kube.Kubernetes.BatchV1().Jobs(namespace).Get(hookName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(newHookObj.UID).NotTo(Equal(hookObj.UID))
		})
	})
})
