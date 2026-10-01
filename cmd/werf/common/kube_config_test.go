package common

import (
	"context"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/kube"
)

var _ = ginkgo.Describe("Synchronization kubeconfig defaults", func() {
	ginkgo.It("loads the default config when legacy kubeconfig flags are empty", func() {
		configData, err := os.ReadFile("testdata/synchronization-kubeconfig.yaml")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		configPath := filepath.Join(ginkgo.GinkgoT().TempDir(), "config")
		gomega.Expect(os.WriteFile(configPath, configData, 0o600)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("KUBECONFIG", configPath)
		data := CmdData{LegacyKubeConfigPathsMergeList: []string{""}}
		gomega.Expect(data.mapLegacyFlags()).To(gomega.Succeed())
		config, err := kube.NewKubeConfig(context.Background(), kube.KubeConfigOptions{KubeConnectionOptions: data.KubeConnectionOptions, KubeContextNamespace: "sync"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(config.RestConfig.Host).To(gomega.Equal("https://sync-cluster.example:6443"))
	})
})
