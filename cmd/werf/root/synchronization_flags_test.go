package root

import (
	"context"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.DescribeTable("Kubernetes synchronization flags", func(commandPath string) {
	ginkgo.GinkgoT().Setenv("WERF_KUBE_CONTEXT", "environment-context")
	ginkgo.GinkgoT().Setenv("WERF_KUBE_CONFIG_BASE64", "ZW52aXJvbm1lbnQ=")
	root, err := ConstructRootCmd(context.Background())
	gomega.Expect(err).To(gomega.Succeed())
	cmd, remaining, err := root.Find(strings.Fields(commandPath))
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(remaining).To(gomega.BeEmpty())
	for _, name := range []string{"kube-context", "kube-config", "kube-config-base64"} {
		gomega.Expect(cmd.PersistentFlags().Lookup(name)).NotTo(gomega.BeNil(), commandPath+" --"+name)
	}
	gomega.Expect(cmd.PersistentFlags().Lookup("kube-context").Value.String()).To(gomega.Equal("environment-context"))
	gomega.Expect(cmd.PersistentFlags().Lookup("kube-config-base64").Value.String()).To(gomega.Equal("ZW52aXJvbm1lbnQ="))
	gomega.Expect(cmd.ParseFlags([]string{"--synchronization=kubernetes://sync", "--kube-context=explicit-context", "--kube-config=/tmp/explicit-config", "--kube-config-base64=ZXhwbGljaXQ="})).To(gomega.Succeed())
	gomega.Expect(cmd.PersistentFlags().Lookup("kube-context").Value.String()).To(gomega.Equal("explicit-context"))
	gomega.Expect(cmd.PersistentFlags().Lookup("kube-config").Value.String()).To(gomega.Equal("/tmp/explicit-config"))
	gomega.Expect(cmd.PersistentFlags().Lookup("kube-config-base64").Value.String()).To(gomega.Equal("ZXhwbGljaXQ="))
},
	ginkgo.Entry("build", "build"),
	ginkgo.Entry("bundle publish", "bundle publish"),
	ginkgo.Entry("compose config", "compose config"),
	ginkgo.Entry("export", "export"),
	ginkgo.Entry("managed-images add", "managed-images add"),
	ginkgo.Entry("managed-images ls", "managed-images ls"),
	ginkgo.Entry("managed-images rm", "managed-images rm"),
	ginkgo.Entry("purge", "purge"),
	ginkgo.Entry("run", "run"),
	ginkgo.Entry("stage image", "stage image"),
)
