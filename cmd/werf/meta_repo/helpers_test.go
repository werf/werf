package meta_repo_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func newProject() string {
	dir := ginkgo.GinkgoT().TempDir()
	for _, name := range []string{"home", "tmp"} {
		gomega.Expect(os.Mkdir(filepath.Join(dir, name), 0o700)).To(gomega.Succeed())
	}
	config, err := os.ReadFile("testdata/werf.yaml")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(os.WriteFile(filepath.Join(dir, "werf.yaml"), config, 0o600)).To(gomega.Succeed())
	for _, args := range [][]string{{"init"}, {"add", "werf.yaml"}, {"-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-m", "fixture"}} {
		command := exec.Command("git", args...)
		command.Dir = dir
		output, err := command.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), string(output))
	}
	return dir
}
