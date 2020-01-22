package releaseserver_test

import (
	"encoding/json"
	"fmt"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repair patches creator", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when resources from release has been changed outside of chart", func() {
		var namespace, projectName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = fmt.Sprintf("%s-dev", projectName)
		})

		AfterEach(func() {
			utils.RunCommand("repair_patches_creator_app1-002", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should generate werf.io/repair-patch annotations on objects which has been changed in cluster and out of sync with the chart configuration", func() {
			Expect(werfDeploy("repair_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "disabled"},
			})).To(Succeed())

		GetAndUpdateMycm1:
			mycm1, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			mycm1.Data = make(map[string]string)
			mycm1.Data["newKey"] = "newValue"
			_, err = kube.Kubernetes.CoreV1().ConfigMaps(namespace).Update(mycm1)
			if errors.IsConflict(err) {
				goto GetAndUpdateMycm1
			}
			Expect(err).NotTo(HaveOccurred())

		GetAndUpdateMydeploy1:
			mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			var replicas int32 = 2
			mydeploy1.Spec.Replicas = &replicas
			_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy1)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy1
			}
			Expect(err).NotTo(HaveOccurred())

			Expect(werfDeploy("repair_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "disabled"},
			})).To(Succeed())

			mycm1, err = kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			d, err := json.Marshal(mycm1.Data)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(d)).To(Equal(`{"newKey":"newValue"}`))
			Expect(mycm1.Annotations["debug.werf.io/repair-patch"]).To(Equal(`{"data":{"aloe":"aloha","moloko":"omlet"}}`))

			_, err = kube.Kubernetes.CoreV1().ConfigMaps(namespace).Patch("mycm1", types.StrategicMergePatchType, []byte(mycm1.Annotations["debug.werf.io/repair-patch"]))
			Expect(err).NotTo(HaveOccurred())

			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(2)))
			Expect(mydeploy1.Annotations["debug.werf.io/repair-patch"]).To(Equal(`{"spec":{"replicas":1}}`))

			Expect(werfDeploy("repair_patches_creator_app1-002", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "disabled"},
			})).To(Succeed())

			mycm1, err = kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			d, err = json.Marshal(mycm1.Data)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(d)).To(Equal(`{"aloe":"aloha","moloko":"omlet","newKey":"newValue"}`))
			Expect(mycm1.Annotations["debug.werf.io/repair-patch"]).To(Equal(`{}`))

			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(2)))
			Expect(mydeploy1.Annotations["debug.werf.io/repair-patch"]).To(Equal(`{}`))
		})
	})
})
