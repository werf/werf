package deploy_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
	"github.com/werf/werf/test/pkg/utils/resourcesfactory"
)

var _ = Describe("Resources adopter", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("when deploying release with resources that already exist in cluster", func() {
		var namespace, projectName, releaseName string

		BeforeEach(func() {
			projectName = utils.ProjectName()
			namespace = projectName
			releaseName = projectName
		})

		AfterEach(func() {
			utils.RunCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
		})

		It("should fail to deploy release when resources already exist; should adopt already existing resources when adoption annotation is set", func() {
			By("Installing release first time without mydeploy2 and mydeploy4")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "resources_adopter_app1-001", "initial commit")

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			By("creating mydeploy2 and mydeploy4 using API in the cluster")

			mydeploy2Original, err := kube.Client.AppsV1().Deployments(namespace).Create(context.Background(), resourcesfactory.NewDeployment(`
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
`), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			mydeploy4Original, err := kube.Client.AppsV1().Deployments(namespace).Create(context.Background(), resourcesfactory.NewDeployment(`
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
`), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("redeploying release with added mydeploy2 and mydeploy4")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "resources_adopter_app1-002", "add mydeploy2 and mydeploy4")

			gotMydeploy2AlreadyExists := false
			gotMydeploy4AlreadyExists := false
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy2" in namespace "%s" exists and cannot be imported into the current release`, namespace)) {
						gotMydeploy2AlreadyExists = true
					}

					if strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy4" in namespace "%s" exists and cannot be imported into the current release`, namespace)) {
						gotMydeploy4AlreadyExists = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotMydeploy2AlreadyExists || gotMydeploy4AlreadyExists).Should(BeTrue())

			mydeploy2AfterDeploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterDeploy.UID).To(Equal(mydeploy2Original.UID))

			mydeploy4AfterDeploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterDeploy.UID).To(Equal(mydeploy4Original.UID))

			By("redeploying release after first failure")

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).NotTo(Succeed())

			By("redeploying release with adoption annotations set to wrong release name")

		GetAndUpdateMydeploy2AfterRedeploy:
			mydeploy2AfterRedeploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterRedeploy.UID).To(Equal(mydeploy2Original.UID))
			mydeploy2AfterRedeploy.Annotations["werf.io/allow-adoption-by-release"] = "NO_SUCH_RELEASE"
			mydeploy2AfterRedeploy, err = kube.Client.AppsV1().Deployments(namespace).Update(context.Background(), mydeploy2AfterRedeploy, metav1.UpdateOptions{})
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy2AfterRedeploy
			}
			Expect(err).NotTo(HaveOccurred())

		GetAndUpdateMydeploy4AfterRedeploy:
			mydeploy4AfterRedeploy, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterRedeploy.UID).To(Equal(mydeploy4Original.UID))
			mydeploy4AfterRedeploy.Annotations["werf.io/allow-adoption-by-release"] = "NO_SUCH_RELEASE"
			mydeploy4AfterRedeploy, err = kube.Client.AppsV1().Deployments(namespace).Update(context.Background(), mydeploy4AfterRedeploy, metav1.UpdateOptions{})
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy4AfterRedeploy
			}
			Expect(err).NotTo(HaveOccurred())

			gotMydeploy2AlreadyExists = false
			gotMydeploy4AlreadyExists = false
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy2" in namespace "%s" exists and cannot be imported into the current release`, namespace)) {
						gotMydeploy2AlreadyExists = true
					}

					if strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy4" in namespace "%s" exists and cannot be imported into the current release`, namespace)) {
						gotMydeploy2AlreadyExists = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotMydeploy2AlreadyExists || gotMydeploy4AlreadyExists).Should(BeTrue())

			By("redeploying release with adoption annotations set to the right release name")

		GetAndUpdateMydeploy2AfterRedeploy2:
			mydeploy2AfterRedeploy, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterRedeploy.UID).To(Equal(mydeploy2Original.UID))
			mydeploy2AfterRedeploy.Labels["app.kubernetes.io/managed-by"] = "Helm"
			mydeploy2AfterRedeploy.Annotations["meta.helm.sh/release-name"] = releaseName
			mydeploy2AfterRedeploy.Annotations["meta.helm.sh/release-namespace"] = namespace
			mydeploy2AfterRedeploy, err = kube.Client.AppsV1().Deployments(namespace).Update(context.Background(), mydeploy2AfterRedeploy, metav1.UpdateOptions{})
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy2AfterRedeploy2
			}
			Expect(err).NotTo(HaveOccurred())

		GetAndUpdateMydeploy4AfterRedeploy2:
			mydeploy4AfterRedeploy, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterRedeploy.UID).To(Equal(mydeploy4Original.UID))
			mydeploy4AfterRedeploy.Labels["app.kubernetes.io/managed-by"] = "Helm"
			mydeploy4AfterRedeploy.Annotations["meta.helm.sh/release-name"] = releaseName
			mydeploy4AfterRedeploy.Annotations["meta.helm.sh/release-namespace"] = namespace
			mydeploy4AfterRedeploy, err = kube.Client.AppsV1().Deployments(namespace).Update(context.Background(), mydeploy4AfterRedeploy, metav1.UpdateOptions{})
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy4AfterRedeploy2
			}
			Expect(err).NotTo(HaveOccurred())

			gotMydeploy2AlreadyExists = false
			gotMydeploy4AlreadyExists = false
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy2" in namespace "%s" exists and cannot be imported into the current release`, namespace))).To(BeFalse(), fmt.Sprintf("Got unexpected output line: %v", line))
					Expect(strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy4" in namespace "%s" exists and cannot be imported into the current release`, namespace))).To(BeFalse(), fmt.Sprintf("Got unexpected output line: %v", line))
				},
			})).To(Succeed())

			mydeploy2AfterAdoption, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy2", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy2AfterAdoption.UID).To(Equal(mydeploy2Original.UID))

			mydeploy4AfterAdoption, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy4", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy4AfterAdoption.UID).To(Equal(mydeploy4Original.UID))

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

			mydeploy5Initial, err := kube.Client.AppsV1().Deployments(namespace).Create(context.Background(), resourcesfactory.NewDeployment(`
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
`), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("updating release with a new resource added to the chart that already exists in the cluster")

			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "resources_adopter_app1-003", "add mydeploy5")

			gotMydeploy5AlreadyExists := false
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy5" in namespace "%s" exists and cannot be imported into the current release`, namespace)) {
						gotMydeploy5AlreadyExists = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotMydeploy5AlreadyExists).To(BeTrue())

		GetAndUpdateMydeploy5AfterUpdate:
			mydeploy5AfterUpdate, err := kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy5", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mydeploy5AfterUpdate.UID).To(Equal(mydeploy5Initial.UID))

			mydeploy5AfterUpdate.Labels["app.kubernetes.io/managed-by"] = "Helm"
			mydeploy5AfterUpdate.Annotations["meta.helm.sh/release-name"] = releaseName
			mydeploy5AfterUpdate.Annotations["meta.helm.sh/release-namespace"] = namespace

			mydeploy5AfterUpdate, err = kube.Client.AppsV1().Deployments(namespace).Update(context.Background(), mydeploy5AfterUpdate, metav1.UpdateOptions{})
			if errors.IsConflict(err) {
				goto GetAndUpdateMydeploy5AfterUpdate
			}
			Expect(err).NotTo(HaveOccurred())

			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					Expect(strings.Contains(line, fmt.Sprintf(`Deployment "mydeploy5" in namespace "%s" exists and cannot be imported into the current release`, namespace))).To(BeFalse(), fmt.Sprintf("Got unexpected output line: %v", line))
				},
			})).To(Succeed())

			By("deleting release from cluster with all adopted resources")

			Expect(liveexec.ExecCommand(SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, liveexec.ExecCommandOptions{}, utils.WerfBinArgs("dismiss")...)).To(Succeed())

			_, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy1", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy1 should return not found error, got %v", err))

			_, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy2", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy2 should return not found error, got %v", err))

			_, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy3", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy3 should return not found error, got %v", err))

			_, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy4", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy4 should return not found error, got %v", err))

			_, err = kube.Client.AppsV1().Deployments(namespace).Get(context.Background(), "mydeploy5", metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue(), fmt.Sprintf("get deploy/mydeploy5 should return not found error, got %v", err))

			kube.Client.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
		})
	})
})
