package e2e_cleanup_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func readCleanupReport(repoPath string) cleanupReport {
	ginkgo.GinkgoHelper()
	data, err := os.ReadFile(filepath.Join(repoPath, ".werf-cleanup-report.json"))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	var report cleanupReport
	gomega.Expect(json.Unmarshal(data, &report)).To(gomega.Succeed())
	return report
}
