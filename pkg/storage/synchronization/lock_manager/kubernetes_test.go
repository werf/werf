package lock_manager

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"

	"github.com/werf/common-go/pkg/locker_with_retry"
	"github.com/werf/nelm/v2/pkg/common"
)

var _ = ginkgo.DescribeTable("Kubernetes synchronization address", func(address string, expected KubernetesParams) {
	actual, err := ParseKubernetesParams(address)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(*actual).To(gomega.Equal(expected))
},
	ginkgo.Entry("namespace", "kubernetes://sync", KubernetesParams{Namespace: "sync"}),
	ginkgo.Entry("context and path", "kubernetes://sync:prod@/tmp/config", KubernetesParams{Namespace: "sync", ConfigContext: "prod", ConfigPath: "/tmp/config"}),
	ginkgo.Entry("embedded config", "kubernetes://sync@base64:YWJj", KubernetesParams{Namespace: "sync", ConfigDataBase64: "YWJj"}),
)

var _ = ginkgo.Describe("Kubernetes synchronization", func() {
	ginkgo.It("rejects other address schemes", func() {
		_, err := ParseKubernetesParams("https://sync")
		gomega.Expect(err).To(gomega.MatchError(`bad address "https://sync": expected kubernetes:// scheme`))
	})
	ginkgo.It("preserves kube flags unless address explicitly overrides them", func() {
		opts := common.KubeConnectionOptions{KubeContextCurrent: "flag-context", KubeConfigPaths: []string{"flag-config"}, KubeConfigBase64: "flag-base64"}
		sync, err := NewKubernetesSynchronization(context.Background(), SynchronizationParams{ServerAddress: "kubernetes://sync", KubeConnectionOptions: opts})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sync.kubeConnectionOptions).To(gomega.Equal(opts))
		sync, err = NewKubernetesSynchronization(context.Background(), SynchronizationParams{ServerAddress: "kubernetes://sync:explicit@/tmp/config", KubeConnectionOptions: opts})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sync.kubeConnectionOptions.KubeContextCurrent).To(gomega.Equal("explicit"))
		gomega.Expect(sync.kubeConnectionOptions.KubeConfigPaths).To(gomega.Equal([]string{"/tmp/config"}))
		gomega.Expect(sync.kubeConnectionOptions.KubeConfigBase64).To(gomega.BeEmpty())
	})
	ginkgo.It("returns the same retry-enabled locker on first and subsequent calls", func() {
		manager := NewKubernetes(context.Background(), "sync", kubernetesfake.NewClientset(), dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()), func(project string) string { return "werf-" + project })
		first, err := manager.getLockerForProject(context.Background(), "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(first).To(gomega.BeAssignableToTypeOf(&locker_with_retry.LockerWithRetry{}))
		second, err := manager.getLockerForProject(context.Background(), "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(second).To(gomega.BeIdenticalTo(first))
	})
})
