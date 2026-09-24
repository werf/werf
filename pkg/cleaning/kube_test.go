package cleaning

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/cmd/werf/common"
)

var _ = ginkgo.Describe("kubernetes namespaces scan", func() {
	ginkgo.DescribeTable("GetKubernetesNamespacesByContext",
		func(scanContextNamespaceOnly bool, kubeScanNamespaces []string, expectedNamespacesByContext map[string][]string) {
			cmdData := &common.CmdData{
				ScanContextNamespaceOnly: &scanContextNamespaceOnly,
				KubeScanNamespaces:       &kubeScanNamespaces,
			}
			contextClients := []*ContextClient{
				{ContextName: "dev", ContextNamespace: "dev-ns"},
				{ContextName: "inClusterContext", ContextNamespace: "runner-ns"},
			}

			gomega.Expect(GetKubernetesNamespacesByContext(cmdData, contextClients)).To(gomega.Equal(expectedNamespacesByContext))
		},
		ginkgo.Entry("scans all namespaces by default",
			false, nil,
			map[string][]string{"dev": nil, "inClusterContext": nil},
		),
		ginkgo.Entry("scans context namespace only when enabled",
			true, nil,
			map[string][]string{"dev": {"dev-ns"}, "inClusterContext": {"runner-ns"}},
		),
		ginkgo.Entry("explicit namespaces are used for every context",
			false, []string{"ns-a", "ns-b"},
			map[string][]string{"dev": {"ns-a", "ns-b"}, "inClusterContext": {"ns-a", "ns-b"}},
		),
		ginkgo.Entry("explicit namespaces take precedence over context namespace only",
			true, []string{"ns-a", "ns-b"},
			map[string][]string{"dev": {"ns-a", "ns-b"}, "inClusterContext": {"ns-a", "ns-b"}},
		),
	)

	ginkgo.It("unions deployed images of every scanned namespace of every context", func() {
		devContextClient := newFakeContextClient("dev", "dev-ns", map[string][]string{
			"ns-a": {"registry/app:a"},
			"ns-b": {"registry/app:b"},
			"ns-c": {"registry/app:c"},
		})
		prodContextClient := newFakeContextClient("prod", "prod-ns", map[string][]string{
			"ns-a": {"registry/app:prod-a"},
		})

		manager := &cleanupManager{
			KubernetesContextClients: []*ContextClient{devContextClient, prodContextClient},
			KubernetesNamespacesByContext: map[string][]string{
				"dev":  {"ns-a", "ns-b"},
				"prod": {"ns-a"},
			},
		}

		deployedImages, err := manager.deployedDockerImages(context.Background())
		gomega.Expect(err).To(gomega.Succeed())

		var names []string
		for _, deployedImage := range deployedImages {
			names = append(names, deployedImage.Name)
		}
		gomega.Expect(names).To(gomega.ConsistOf("registry/app:a", "registry/app:b", "registry/app:prod-a"))
	})

	ginkgo.It("fails instead of skipping a namespace that cannot be scanned", func() {
		contextClient := newFakeContextClient("dev", "dev-ns", map[string][]string{
			"ns-a": {"registry/app:a"},
			"ns-b": {"registry/app:b"},
		})
		failPodsListInNamespace(contextClient, "ns-b")

		manager := &cleanupManager{
			KubernetesContextClients:      []*ContextClient{contextClient},
			KubernetesNamespacesByContext: map[string][]string{"dev": {"ns-a", "ns-b"}},
		}

		deployedImages, err := manager.deployedDockerImages(context.Background())
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(`pods is forbidden in namespace "ns-b"`)))
		gomega.Expect(deployedImages).To(gomega.BeNil())
	})
})
