package releaseserver_test

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"
	"github.com/flant/werf/pkg/testing/utils/resourcesfactory"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = XDescribe("Resources adopter", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when installing and updating release with resources that already exist in cluster", func() {
		var namespace, projectName, releaseName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = fmt.Sprintf("%s-dev", projectName)
			releaseName = fmt.Sprintf("%s-dev", projectName)
		})

		AfterEach(func() {
			utils.RunCommand("resources_adopter_app1-002", werfBinPath, "dismiss", "--env", "dev", "--with-namespace")
		})

		It("should fail to install release; should not delete already existing resources on failed release removal when reinstalling release; should delete new resources created during failed release installation when reinstalling release; should adopt already existing resources by annotation", func() {
			By("creating mydeploy2 and mydeploy4 using API in the cluster before installing release")

			_, err := kube.Kubernetes.CoreV1().Namespaces().Create(resourcesfactory.NewNamespace(fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace)))
			Expect(err).NotTo(HaveOccurred())

			mydeploy2BeforeInstall, err := kube.Kubernetes.AppsV1().Deployments(namespace).Create(resourcesfactory.NewDeployment(`
kind: Deployment
apiVersion: apps/v1
metadata:
  name: mydeploy2
  labels:
    service: mydeploy2
spec:
  replicas: 2
  selector:
    matchLabels:
      service: mydeploy2
  template:
    metadata:
      labels:
        service: mydeploy2
    spec:
      containers:
      - name: mycontainer1
        command: [ "/bin/bash", "-c", "while true; do date; sleep 1; done" ]
        image: ubuntu:18.04
      - name: mycontainer2
        command: [ "/bin/bash", "-c", "while true; do date; sleep 1; done" ]
        image: ubuntu:18.04
`))
			Expect(err).NotTo(HaveOccurred())

			mydeploy4BeforeInstall, err := kube.Kubernetes.AppsV1().Deployments(namespace).Create(resourcesfactory.NewDeployment(`
kind: Deployment
apiVersion: apps/v1
metadata:
  name: mydeploy4
  annotations:
    alo: alo
  labels:
    service: mydeploy4
    helo: world
spec:
  replicas: 2
  selector:
    matchLabels:
      service: mydeploy4
  template:
    metadata:
      labels:
        service: mydeploy4
        helo: world
    spec:
      containers:
      - name: main
        command: [ "/bin/bash", "-c", "while true; do date; sleep 1; done" ]
        image: ubuntu:18.04
        env:
        - name: MYVAR
          value: anotherValue
        - name: MYVAR2
          value: "123"
`))
			Expect(err).NotTo(HaveOccurred())

			By("installing release first time")

			gotMydeploy2AlreadyExists := false
			gotMydeploy4AlreadyExists := false
			Expect(werfDeploy("resources_adopter_app1-001", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, "Deployment/mydeploy2 already exists in the cluster") != -1 {
						gotMydeploy2AlreadyExists = true
					}
					if strings.Index(line, "Deployment/mydeploy4 already exists in the cluster") != -1 {
						gotMydeploy4AlreadyExists = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotMydeploy2AlreadyExists || gotMydeploy4AlreadyExists).Should(BeTrue())

			for {
				_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
				if err == nil {
					time.Sleep(200 * time.Millisecond)
				} else if errors.IsNotFound(err) {
					break
				} else {
					Fail(fmt.Sprintf("error accessing deploy/mydeploy1: %s", err))
				}
			}

			for {
				_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy3", metav1.GetOptions{})
				if err == nil {
					time.Sleep(200 * time.Millisecond)
				} else if errors.IsNotFound(err) {
					break
				} else {
					Fail(fmt.Sprintf("error accessing deploy/mydeploy3: %s", err))
				}
			}

			mydeploy2AfterInstall, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterInstall.UID).To(Equal(mydeploy2BeforeInstall.UID))

			mydeploy4AfterInstall, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterInstall.UID).To(Equal(mydeploy4BeforeInstall.UID))

			By("reinstalling release after first failure")

			Expect(werfDeploy("resources_adopter_app1-001", liveexec.ExecCommandOptions{})).NotTo(Succeed())

			By("reinstalling release with adoption annotations set to wrong release name")

		GetAndUpdateMydeploy2AfterReinstall:
			mydeploy2AfterReinstall, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterReinstall.UID).To(Equal(mydeploy2BeforeInstall.UID))
			mydeploy2AfterReinstall.Annotations["werf.io/allow-adoption-by-release"] = "NO_SUCH_RELEASE"
			mydeploy2AfterReinstall, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy2AfterReinstall)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy2AfterReinstall
			}
			Expect(err).NotTo(HaveOccurred())

		GetAndUpdateMydeploy4AfterReinstall:
			mydeploy4AfterReinstall, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterReinstall.UID).To(Equal(mydeploy4BeforeInstall.UID))
			mydeploy4AfterReinstall.Annotations["werf.io/allow-adoption-by-release"] = "NO_SUCH_RELEASE"
			mydeploy4AfterReinstall, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy4AfterReinstall)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy4AfterReinstall
			}
			Expect(err).NotTo(HaveOccurred())

			gotMydeploy2AlreadyExists = false
			gotMydeploy4AlreadyExists = false
			Expect(werfDeploy("resources_adopter_app1-001", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, "Deployment/mydeploy2 already exists in the cluster") != -1 {
						gotMydeploy2AlreadyExists = true
					}
					if strings.Index(line, "Deployment/mydeploy4 already exists in the cluster") != -1 {
						gotMydeploy4AlreadyExists = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotMydeploy2AlreadyExists || gotMydeploy4AlreadyExists).Should(BeTrue())

			By("reinstalling release with adoption annotations set to the right release name")

		GetAndUpdateMydeploy2AfterReinstall2:
			mydeploy2AfterReinstall, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterReinstall.UID).To(Equal(mydeploy2BeforeInstall.UID))
			mydeploy2AfterReinstall.Annotations["werf.io/allow-adoption-by-release"] = releaseName
			mydeploy2AfterReinstall, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy2AfterReinstall)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy2AfterReinstall2
			}
			Expect(err).NotTo(HaveOccurred())

		GetAndUpdateMydeploy4AfterReinstall2:
			mydeploy4AfterReinstall, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterReinstall.UID).To(Equal(mydeploy4BeforeInstall.UID))
			mydeploy4AfterReinstall.Annotations["werf.io/allow-adoption-by-release"] = releaseName
			mydeploy4AfterReinstall, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy4AfterReinstall)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy4AfterReinstall2
			}
			Expect(err).NotTo(HaveOccurred())

			gotMydeploy2AlreadyExists = false
			gotMydeploy4AlreadyExists = false
			Expect(werfDeploy("resources_adopter_app1-001", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.Index(line, "Deployment/mydeploy2 already exists in the cluster")).To(Equal(-1), fmt.Sprintf("Got unexpected output line: %v", line))
					Expect(strings.Index(line, "Deployment/mydeploy4 already exists in the cluster")).To(Equal(-1), fmt.Sprintf("Got unexpected output line: %v", line))
				},
			})).To(Succeed())

			mydeploy2AfterAdoption, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterAdoption.UID).To(Equal(mydeploy2BeforeInstall.UID))

			mydeploy4AfterAdoption, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterAdoption.UID).To(Equal(mydeploy4BeforeInstall.UID))

			Expect(mydeploy2AfterAdoption.Annotations["service.werf.io/owner-release"]).To(Equal(releaseName))
			Expect(mydeploy4AfterAdoption.Annotations["service.werf.io/owner-release"]).To(Equal(releaseName))

			Expect(*mydeploy2AfterAdoption.Spec.Replicas).To(Equal(int32(1)))

			Expect(len(mydeploy2AfterAdoption.Spec.Template.Spec.Containers)).To(Equal(3))
			mainContainerFound := false
			for _, containerSpec := range mydeploy2AfterAdoption.Spec.Template.Spec.Containers {
				if containerSpec.Name == "main" {
					mainContainerFound = true
					Expect(len(containerSpec.Env)).To(Equal(1))
					Expect(containerSpec.Env[0].Name).To(Equal("KEY"))
					Expect(containerSpec.Env[0].Value).To(Equal("VALUE"))
				}
			}
			Expect(mainContainerFound).To(BeTrue())

			Expect(mydeploy4AfterAdoption.Annotations["alo"]).To(Equal("alo"))
			Expect(mydeploy4AfterAdoption.Labels["helo"]).To(Equal("world"))
			Expect(*mydeploy4AfterAdoption.Spec.Replicas).To(Equal(int32(1)))
			Expect(mydeploy4AfterAdoption.Spec.Template.Labels["helo"]).To(Equal("world"))

			Expect(len(mydeploy4AfterAdoption.Spec.Template.Spec.Containers[0].Env)).To(Equal(2))
			myvarFound := false
			for _, envSpec := range mydeploy4AfterAdoption.Spec.Template.Spec.Containers[0].Env {
				if envSpec.Name == "MYVAR" {
					myvarFound = true
					Expect(envSpec.Value).To(Equal("myvalue"))
				}
			}
			Expect(myvarFound).To(BeTrue())

			By("creating mydeploy5 in the cluster using API")

			mydeploy5Initial, err := kube.Kubernetes.AppsV1().Deployments(namespace).Create(resourcesfactory.NewDeployment(`
kind: Deployment
apiVersion: apps/v1
metadata:
  name: mydeploy5
  labels:
    service: mydeploy5
spec:
  replicas: 2
  selector:
    matchLabels:
      service: mydeploy5
  template:
    metadata:
      labels:
        service: mydeploy5
    spec:
      containers:
      - name: main
        command: [ "/bin/bash", "-c", "while true; do date; sleep 1; done" ]
        image: ubuntu:18.04
`))
			Expect(err).NotTo(HaveOccurred())

			By("updating release with a new resource added to the chart that already exists in the cluster")

			gotMydeploy5AlreadyExists := false
			Expect(werfDeploy("resources_adopter_app1-002", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, "Deployment/mydeploy5 already exists in the cluster") != -1 {
						gotMydeploy5AlreadyExists = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotMydeploy5AlreadyExists).To(BeTrue())

		GetAndUpdateMydeploy5AfterUpdate:
			mydeploy5AfterUpdate, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy5", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy5AfterUpdate.UID).To(Equal(mydeploy5Initial.UID))
			mydeploy5AfterUpdate.Annotations["werf.io/allow-adoption-by-release"] = releaseName
			mydeploy5AfterUpdate, err = kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy5AfterUpdate)
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy5AfterUpdate
			}
			Expect(err).NotTo(HaveOccurred())

			Expect(werfDeploy("resources_adopter_app1-002", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.Index(line, "Deployment/mydeploy5 already exists in the cluster")).To(Equal(-1), fmt.Sprintf("Got unexpected output line: %v", line))
				},
			})).To(Succeed())

			By("deleting release from cluster with all adopted resources")

			Expect(werfDismiss("resources_adopter_app1-002", liveexec.ExecCommandOptions{})).To(Succeed())

			_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy1 should return not found error, got %v", err))

			_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy2", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy2 should return not found error, got %v", err))

			_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy3", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy3 should return not found error, got %v", err))

			_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy4", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy4 should return not found error, got %v", err))

			_, err = kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy5", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy5 should return not found error, got %v", err))

			_, err = kube.Kubernetes.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get ns/%s should return not found error, got %v", namespace, err))
		})
	})
})
