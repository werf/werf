package chart_ts_init

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func initProject() string {
	ginkgo.GinkgoHelper()
	dir := ginkgo.GinkgoT().TempDir()
	gomega.Expect(os.WriteFile(filepath.Join(dir, "werf.yaml"), []byte("configVersion: 1\nproject: scaffold-test\n"), 0o644)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}\n"), 0o644)).To(gomega.Succeed())
	for _, args := range [][]string{{"init"}, {"add", "."}, {"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-m", "fixture"}} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		output, err := cmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s", output)
	}
	return dir
}

func readScaffoldFile(tsDir, name string) string {
	ginkgo.GinkgoHelper()
	content, err := os.ReadFile(filepath.Join(tsDir, filepath.FromSlash(name)))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return string(content)
}
