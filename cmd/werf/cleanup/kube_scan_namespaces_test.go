package cleanup

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/cmd/werf/common"
)

var _ = ginkgo.Describe("cleanup command --kube-scan-namespaces", func() {
	ginkgo.It("registers the flag", func() {
		flag := NewCmd(context.Background()).Flags().Lookup("kube-scan-namespaces")

		gomega.Expect(flag).NotTo(gomega.BeNil())
		gomega.Expect(flag.Value.Type()).To(gomega.Equal("stringArray"))
	})

	ginkgo.It("collects repeated flag values into cmd data", func() {
		cmd := NewCmd(context.Background())

		gomega.Expect(cmd.ParseFlags([]string{"--kube-scan-namespaces", "ns-a", "--kube-scan-namespaces", "ns-b"})).To(gomega.Succeed())

		gomega.Expect(*commonCmdData.KubeScanNamespaces).To(gomega.Equal([]string{"ns-a", "ns-b"}))
	})

	ginkgo.It("combines environment values with repeated flag values", func() {
		ginkgo.GinkgoT().Setenv("WERF_KUBE_SCAN_NAMESPACES_1", "env-ns-1")
		ginkgo.GinkgoT().Setenv("WERF_KUBE_SCAN_NAMESPACES_2", "env-ns-2")

		cmd := NewCmd(context.Background())

		gomega.Expect(cmd.ParseFlags([]string{"--kube-scan-namespaces", "cli-ns"})).To(gomega.Succeed())

		gomega.Expect(common.GetKubeScanNamespaces(&commonCmdData)).To(gomega.Equal([]string{"env-ns-1", "env-ns-2", "cli-ns"}))
	})
})
