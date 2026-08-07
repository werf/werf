package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/kubedog/pkg/kube"
)

var _ = Describe("kubernetes namespaces scan", func() {
	It("uses explicit namespaces when provided", func() {
		scanContextNamespaceOnly := false
		scanNamespaces := []string{"ns-a", "ns-b"}
		cmdData := &CmdData{
			ScanContextNamespaceOnly: &scanContextNamespaceOnly,
			KubeScanNamespaces:       &scanNamespaces,
		}
		contextClients := []*kube.ContextClient{
			{ContextName: "dev", ContextNamespace: "dev-ns"},
		}

		res := GetKubernetesNamespacesByContext(cmdData, contextClients)

		Expect(res["dev"]).To(Equal([]string{"ns-a", "ns-b"}))
	})

	It("uses context namespace when scan-context-namespace-only is enabled", func() {
		scanContextNamespaceOnly := true
		scanNamespaces := []string{}
		cmdData := &CmdData{
			ScanContextNamespaceOnly: &scanContextNamespaceOnly,
			KubeScanNamespaces:       &scanNamespaces,
		}
		contextClients := []*kube.ContextClient{
			{ContextName: "dev", ContextNamespace: "dev-ns"},
		}

		res := GetKubernetesNamespacesByContext(cmdData, contextClients)

		Expect(res["dev"]).To(Equal([]string{"dev-ns"}))
	})

	It("scans all namespaces by default for kubeconfig contexts", func() {
		scanContextNamespaceOnly := false
		scanNamespaces := []string{}
		cmdData := &CmdData{
			ScanContextNamespaceOnly: &scanContextNamespaceOnly,
			KubeScanNamespaces:       &scanNamespaces,
		}
		contextClients := []*kube.ContextClient{
			{ContextName: "dev", ContextNamespace: "dev-ns"},
		}

		res := GetKubernetesNamespacesByContext(cmdData, contextClients)

		Expect(res["dev"]).To(BeNil())
	})

	It("scans all namespaces by default for in-cluster context", func() {
		scanContextNamespaceOnly := false
		scanNamespaces := []string{}
		cmdData := &CmdData{
			ScanContextNamespaceOnly: &scanContextNamespaceOnly,
			KubeScanNamespaces:       &scanNamespaces,
		}
		contextClients := []*kube.ContextClient{
			{ContextName: "inClusterContext", ContextNamespace: "runner-ns"},
		}

		res := GetKubernetesNamespacesByContext(cmdData, contextClients)

		Expect(res["inClusterContext"]).To(BeNil())
	})
})
