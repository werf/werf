package e2e_helm_secret_test

import (
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

func generateSecretKey(ctx ginkgo.SpecContext) string {
	ginkgo.GinkgoHelper()
	return strings.TrimSpace(utils.SucceedCommandOutputString(ctx, "", SuiteData.WerfBinPath, "helm", "secret", "generate-secret-key"))
}

func readFile(path string) string {
	ginkgo.GinkgoHelper()
	data, err := os.ReadFile(path)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return string(data)
}
