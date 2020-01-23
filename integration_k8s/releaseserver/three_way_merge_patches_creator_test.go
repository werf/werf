package releaseserver_test

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Three way merge patches creator", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when deploying an existing release with the onlyNewReleases 3wm mode", func() {
		AfterEach(func() {
			utils.RunCommand("three_way_merge_patches_creator_app1-001", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should not use 3wm patches during release update", func() {
			By("initial release installation with disabled 3wm")
			gotCorrect3wmModeLine := false
			got3wmDisabledLine := false
			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "disabled"},
				OutputLineHandler: func(line string) {
					if strings.Index(line, `Using three-way-merge mode "disabled"`) != -1 {
						gotCorrect3wmModeLine = true
					}
					if strings.Index(line, `Three way merge is DISABLED for the release`) != -1 {
						got3wmDisabledLine = true
					}
				},
			})).To(Succeed())
			Expect(gotCorrect3wmModeLine).To(BeTrue())
			Expect(got3wmDisabledLine).To(BeTrue())

			By("release update with onlyNewReleases 3wm mode")
			gotCorrect3wmModeLine = false
			got3wmDisabledLine = false
			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "onlyNewReleases"},
				OutputLineHandler: func(line string) {
					if strings.Index(line, `Using three-way-merge mode "onlyNewReleases"`) != -1 {
						gotCorrect3wmModeLine = true
					}
					if strings.Index(line, `Three way merge is DISABLED for the release`) != -1 {
						got3wmDisabledLine = true
					}
				},
			})).To(Succeed())
			Expect(gotCorrect3wmModeLine).To(BeTrue())
			Expect(got3wmDisabledLine).To(BeTrue())
		})
	})

	Context("when deploying a new release with the onlyNewReleases 3wm mode", func() {
		AfterEach(func() {
			utils.RunCommand("three_way_merge_patches_creator_app1-001", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should use 3wm patches during release installation and subsequent update", func() {
			By("initial release installation with onlyNewReleases 3wm mode")
			gotCorrect3wmModeLine := false
			got3wmEnabledLine := false
			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "onlyNewReleases"},
				OutputLineHandler: func(line string) {
					if strings.Index(line, `Using three-way-merge mode "onlyNewReleases"`) != -1 {
						gotCorrect3wmModeLine = true
					}
					if strings.Index(line, `Three way merge is ENABLED for the release`) != -1 {
						got3wmEnabledLine = true
					}
				},
			})).To(Succeed())
			Expect(gotCorrect3wmModeLine).To(BeTrue())
			Expect(got3wmEnabledLine).To(BeTrue())

			By("release update with onlyNewReleases 3wm mode")
			gotCorrect3wmModeLine = false
			got3wmEnabledLine = false
			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "onlyNewReleases"},
				OutputLineHandler: func(line string) {
					if strings.Index(line, `Using three-way-merge mode "onlyNewReleases"`) != -1 {
						gotCorrect3wmModeLine = true
					}
					if strings.Index(line, `Three way merge is ENABLED for the release`) != -1 {
						got3wmEnabledLine = true
					}
				},
			})).To(Succeed())
			Expect(gotCorrect3wmModeLine).To(BeTrue())
			Expect(got3wmEnabledLine).To(BeTrue())
		})
	})

	Context("when deploying an existing release with the enabled 3wm mode", func() {
		AfterEach(func() {
			utils.RunCommand("three_way_merge_patches_creator_app1-001", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should use 3wm patches during release update", func() {
			By("initial release installation with disabled 3wm mode")
			gotCorrect3wmModeLine := false
			got3wmDisabledLine := false
			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "disabled"},
				OutputLineHandler: func(line string) {
					if strings.Index(line, `Using three-way-merge mode "disabled"`) != -1 {
						gotCorrect3wmModeLine = true
					}
					if strings.Index(line, `Three way merge is DISABLED for the release`) != -1 {
						got3wmDisabledLine = true
					}
				},
			})).To(Succeed())
			Expect(gotCorrect3wmModeLine).To(BeTrue())
			Expect(got3wmDisabledLine).To(BeTrue())

			By("release update with enabled 3wm mode")
			gotCorrect3wmModeLine = false
			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
				OutputLineHandler: func(line string) {
					if strings.Index(line, `Using three-way-merge mode "enabled"`) != -1 {
						gotCorrect3wmModeLine = true
					}
					Expect(line).NotTo(ContainSubstring(`Three way merge is DISABLED for the release`))
				},
			})).To(Succeed())
			Expect(gotCorrect3wmModeLine).To(BeTrue())
		})
	})

	Context("when release resources has been changed manually and 3wm is enabled", func() {
		var namespace, projectName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = fmt.Sprintf("%s-dev", projectName)
		})

		AfterEach(func() {
			utils.RunCommand("three_way_merge_patches_creator_app1-001", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should bring resources to the chart state with 3wm paches", func() {
			By("initial release installation")

			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			By("changing release resources manually")

			changeResourcesManually := func() {
			GetAndUpdateMydeploy1:
				mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				for _, c := range mydeploy1.Spec.Template.Spec.Containers {
					c.Image = "alpine"
				}
				mydeploy1.Spec.Replicas = new(int32)
				*mydeploy1.Spec.Replicas = 2
				_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy1)
				if errors.IsConflict(err) {
					goto GetAndUpdateMydeploy1
				}
				Expect(err).NotTo(HaveOccurred())

			GetAndUpdateMycm1:
				mycm1, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				delete(mycm1.Data, "moloko")
				mycm1.Data["aloe"] = "cactus"
				mycm1.Data["testo"] = "cheburek"
				mycm1.Annotations["extraKey"] = "value"
				mycm1.Labels = make(map[string]string)
				mycm1.Labels["extraKey"] = "value"
				_, err = kube.Kubernetes.CoreV1().ConfigMaps(namespace).Update(mycm1)
				if errors.IsConflict(err) {
					goto GetAndUpdateMycm1
				}
				Expect(err).NotTo(HaveOccurred())
			}

			changeResourcesManually()

			By("redeploying same chart")

			Expect(werfDeploy("three_way_merge_patches_creator_app1-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(1)))
			Expect(len(mydeploy1.Spec.Template.Spec.Containers)).To(Equal(1))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Image).To(Equal("ubuntu:18.04"))

			mycm1, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mycm1.Annotations["extraKey"]).To(Equal("value"))
			Expect(mycm1.Labels["extraKey"]).To(Equal("value"))
			Expect(mycm1.Data["aloe"]).To(Equal("aloha"))
			Expect(mycm1.Data["moloko"]).To(Equal("omlet"))
			Expect(mycm1.Data["testo"]).To(Equal("cheburek"))

			By("change release resources manually 2-nd time")

			changeResourcesManually()

			By("redeploying chart with changes since previous chart version")

			Expect(werfDeploy("three_way_merge_patches_creator_app1-002", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(3)))
			Expect(len(mydeploy1.Spec.Template.Spec.Containers)).To(Equal(1))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Name).To(Equal("main2"))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Image).To(Equal("ubuntu:19.04"))

			mycm1, err = kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mycm1.Annotations["extraKey"]).To(Equal("value"))
			Expect(mycm1.Labels["extraKey"]).To(Equal("value"))

			Expect(mycm1.Data["aloe"]).To(BeEmpty())
			Expect(mycm1.Data["moloko"]).To(BeEmpty())
			Expect(mycm1.Data["testo"]).To(Equal("bulka"))
			Expect(mycm1.Data["one"]).To(Equal("1"))
			Expect(mycm1.Data["two"]).To(Equal("2"))
			Expect(mycm1.Data["three"]).To(Equal("3"))

			Expect(werfDeploy("three_way_merge_patches_creator_app1-002", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())
		})
	})

	Context("when release resources and replicas has been changed manually and 3wm is enabled and set-replicas/resources-only-on-creation was not set initially", func() {
		var namespace, projectName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = fmt.Sprintf("%s-dev", projectName)
		})

		AfterEach(func() {
			utils.RunCommand("three_way_merge_patches_creator_app2-002", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should stop changing replicas and resources on redeploy when user sets set-replicas/resources-only-on-creation annotation to the resource", func() {
			By("deploying chart initially")

			Expect(werfDeploy("three_way_merge_patches_creator_app2-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			By("chaning resources and replicas manually")

			changeResourcesManually := func() {
				mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				mydeploy1.Spec.Replicas = new(int32)
				*mydeploy1.Spec.Replicas = 2

				requestCpu := resource.Quantity{Format: resource.DecimalSI}
				requestCpu.SetMilli(30)
				requestMem := resource.Quantity{Format: resource.BinarySI}
				requestMem.Set(128 * 1024 * 1024)

				limitCpu := resource.Quantity{Format: resource.DecimalSI}
				limitCpu.SetMilli(30)
				limitMem := resource.Quantity{Format: resource.BinarySI}
				limitMem.Set(256 * 1024 * 1024)

				mydeploy1.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    requestCpu,
						corev1.ResourceMemory: requestMem,
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    limitCpu,
						corev1.ResourceMemory: limitMem,
					},
				}

				_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy1)
				Expect(err).NotTo(HaveOccurred())

				mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(2)))
				Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()).To(Equal(int64(30)))
				Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Value()).To(Equal(int64(128 * 1024 * 1024)))
				Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(Equal(int64(30)))
				Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value()).To(Equal(int64(256 * 1024 * 1024)))
			}

			changeResourcesManually()

			By("redeploying same chart")

			Expect(werfDeploy("three_way_merge_patches_creator_app2-001", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(1)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()).To(Equal(int64(10)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Value()).To(Equal(int64(64 * 1024 * 1024)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(Equal(int64(10)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value()).To(Equal(int64(128 * 1024 * 1024)))

			By("chaning resources and replicas manually")

			changeResourcesManually()

			By("redeploying chart with set-replicas/resources-only-on-creation annotations")

			Expect(werfDeploy("three_way_merge_patches_creator_app2-002", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(2)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()).To(Equal(int64(30)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Value()).To(Equal(int64(128 * 1024 * 1024)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(Equal(int64(30)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value()).To(Equal(int64(256 * 1024 * 1024)))

			By("redeploying chart with set-replicas/resources-only-on-creation annotations and changes replicas and resources")

			Expect(werfDeploy("three_way_merge_patches_creator_app2-003", liveexec.ExecCommandOptions{
				Env: map[string]string{"WERF_THREE_WAY_MERGE_MODE": "enabled"},
			})).To(Succeed())

			mydeploy1, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(*mydeploy1.Spec.Replicas).To(Equal(int32(2)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()).To(Equal(int64(30)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().Value()).To(Equal(int64(128 * 1024 * 1024)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(Equal(int64(30)))
			Expect(mydeploy1.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value()).To(Equal(int64(256 * 1024 * 1024)))
		})
	})
})
