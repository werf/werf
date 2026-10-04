package includes

import (
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/werf"
)

func TestIncludesGC(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Includes GC")
}

var _ = ginkgo.Describe("Includes cache eviction", func() {
	ginkgo.It("restores a missing mirror before reading locked includes", func(ctx ginkgo.SpecContext) {
		cfg, repos, commit := setupRemoteInclude(ctx)
		info, err := getLockInfo(ctx, getLockInfoOptions{includesConfig: cfg, useLatestVersion: true, remoteRepos: repos})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		locked, err := info.GetCommit(cfg.Includes[0].Git, "main")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(locked).To(gomega.Equal(commit))
		repo, err := repos.getRepository(cfg.Includes[0].Git)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		remote, ok := repo.repo.(*git_repo.Remote)
		gomega.Expect(ok).To(gomega.BeTrue())
		gomega.Expect(os.RemoveAll(remote.GetClonePath())).To(gomega.Succeed())
		includes, err := GetIncludes(ctx, cfg, info, repos)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(includes).To(gomega.HaveLen(1))
		gomega.Expect(includes[0].commitHash).To(gomega.Equal(commit))
		data, err := includes[0].GetFile(ctx, "data.txt")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(string(data)).To(gomega.Equal("v1"))
	})

	ginkgo.It("keeps GC outside the complete repository callback", func(ctx ginkgo.SpecContext) {
		cfg, repos, _ := setupRemoteInclude(ctx)
		repo, err := repos.getRepository(cfg.Includes[0].Git)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		attempted := make(chan struct{})
		finished := make(chan error, 1)
		err = repo.withRepository(ctx, "", func(handle *git.Repository) error {
			go func() {
				close(attempted)
				lock, err := git_repo.CommonGitDataManager.LockGC(ctx, false)
				if err != nil {
					finished <- err
					return
				}
				finished <- werf.HostLocker().ReleaseLock(lock)
			}()
			gomega.Eventually(attempted).Should(gomega.BeClosed())
			gomega.Consistently(finished, 300*time.Millisecond).ShouldNot(gomega.Receive())
			_, err := handle.Head()
			return err
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Eventually(finished, 5*time.Second).Should(gomega.Receive(gomega.Succeed()))
	})
})
