package true_git_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/true_git/status"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("LFS public status regression", func() {
	ginkgo.It("keeps touched native LFS contents clean but reports real edits", func(ctx ginkgo.SpecContext) {
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "0")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_PARAMETERS", "")
		repoDir := ginkgo.GinkgoT().TempDir()
		utils.RunSucceedCommand(ctx, repoDir, "git", "init", "--initial-branch=main")
		utils.RunSucceedCommand(ctx, repoDir, "git", "lfs", "install", "--local", "--skip-repo")
		utils.RunSucceedCommand(ctx, repoDir, "git", "config", "filter.lfs.clean", "")
		const content = "native process-only LFS content\x00\xff"
		gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitattributes"), []byte("*.bin filter=lfs diff=lfs merge=lfs -text\n"), 0o644)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(filepath.Join(repoDir, "payload.bin"), []byte(content), 0o644)).To(gomega.Succeed())
		utils.RunSucceedCommand(ctx, repoDir, "git", "add", ".")
		utils.RunSucceedCommand(ctx, repoDir, "git", "-c", "user.name=LFS test", "-c", "user.email=lfs@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "native pointer")
		pointer := fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%x\nsize %d\n", sha256.Sum256([]byte(content)), len(content))
		gomega.Expect(utils.SucceedCommandOutputString(ctx, repoDir, "git", "show", "HEAD:payload.bin")).To(gomega.Equal(pointer))

		touched := time.Now().Add(-time.Hour)
		gomega.Expect(os.Chtimes(filepath.Join(repoDir, "payload.bin"), touched, touched)).To(gomega.Succeed())
		result, err := status.Status(ctx, repoDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(result.PathListWithSubmodules()).To(gomega.BeEmpty())

		gomega.Expect(os.WriteFile(filepath.Join(repoDir, "payload.bin"), []byte(content+"changed"), 0o644)).To(gomega.Succeed())
		result, err = status.Status(ctx, repoDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(result.Worktree.PathList()).To(gomega.Equal([]string{"payload.bin"}))
		gomega.Expect(result.Index.PathList()).To(gomega.BeEmpty())
	})
})
