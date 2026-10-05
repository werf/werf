package lock_manager

import (
	"errors"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	kubernetesscheme "k8s.io/client-go/kubernetes/scheme"
	kubetesting "k8s.io/client-go/testing"

	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = ginkgo.DescribeTable("storage lock acquisition statistics",
	func(ctx ginkgo.SpecContext, kubernetes, fail bool) {
		collector := opstats.NewCollector()
		observedCtx := opstats.NewContext(ctx, collector)
		locker := &observedTestLocker{onAcquire: func() {
			gomega.Expect(collector.Summary()).To(gomega.BeEmpty())
		}}
		if fail {
			locker.err = errors.New("acquisition failed")
		}

		var manager Interface = NewGeneric(locker)
		if kubernetes {
			manager = &Kubernetes{LockerPerProject: map[string]lockgate.Locker{"project": locker}}
		}
		handle, err := manager.LockStage(observedCtx, "project", "digest")
		if fail {
			gomega.Expect(err).To(gomega.MatchError(locker.err))
		} else {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}

		summary := collector.Summary()
		gomega.Expect(summary).To(gomega.HaveLen(1))
		gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.Operation("sync: lock acquire")))
		gomega.Expect(summary[0].Count).To(gomega.Equal(1))
		if !fail {
			gomega.Expect(manager.Unlock(observedCtx, handle)).To(gomega.Succeed())
			gomega.Expect(collector.Summary()).To(gomega.Equal(summary))
		}
	},
	ginkgo.Entry("HTTP/local lock success", false, false),
	ginkgo.Entry("HTTP/local lock failure", false, true),
	ginkgo.Entry("Kubernetes lock success", true, false),
	ginkgo.Entry("Kubernetes lock failure", true, true),
)

var _ = ginkgo.Describe("Kubernetes storage lock acquisition statistics", func() {
	const configMapSetupDelay = 300 * time.Millisecond

	ginkgo.It("excludes the ConfigMap setup done by the first acquisition of a project", func(ctx ginkgo.SpecContext) {
		const namespace = "sync"
		const configMapName = "werf-project"

		kubeClient := kubernetesfake.NewSimpleClientset()
		kubeClient.PrependReactor("*", "configmaps", func(kubetesting.Action) (bool, runtime.Object, error) {
			time.Sleep(configMapSetupDelay)
			return false, nil, nil
		})
		kubeDynamicClient := dynamicfake.NewSimpleDynamicClient(kubernetesscheme.Scheme, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: namespace},
		})

		manager := NewKubernetes(namespace, kubeClient, kubeDynamicClient, func(string) string { return configMapName })

		collector := opstats.NewCollector()
		handle, err := manager.LockStage(opstats.NewContext(ctx, collector), "project", "digest")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(func() { gomega.Expect(manager.Unlock(ctx, handle)).To(gomega.Succeed()) })

		summary := collector.Summary()
		gomega.Expect(summary).To(gomega.HaveLen(1))
		gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.Operation("sync: lock acquire")))
		// The setup path sleeps twice (get, then create); the acquisition itself only needs
		// headroom for a scheduler pause, so anything below one delay plus that headroom
		// proves the timer skipped the setup.
		gomega.Expect(summary[0].TotalTime).To(gomega.BeNumerically("<", configMapSetupDelay*3/2), "the timer must not cover creating the locking ConfigMap")
	})
})
