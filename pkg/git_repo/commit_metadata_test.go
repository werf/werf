package git_repo

import (
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = ginkgo.Describe("Commit metadata", func() {
	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		gomega.Expect(Init(&fakeGitDataManager{})).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("reads the requested commit and committer date independently of HEAD and author",
		func(ctx ginkgo.SpecContext, remoteSource bool) {
			dir := newChecksumTestRepo(ctx)
			gomega.Expect(os.WriteFile(filepath.Join(dir, "source"), []byte("unchanged"), 0o644)).To(gomega.Succeed())
			ginkgo.GinkgoT().Setenv("GIT_AUTHOR_DATE", "2001-01-01T00:00:00Z")
			ginkgo.GinkgoT().Setenv("GIT_COMMITTER_DATE", "2024-02-03T12:34:56+03:00")
			commitChecksumTestRepo(ctx, dir)
			first := utils.GetHeadCommit(ctx, dir)
			ginkgo.GinkgoT().Setenv("GIT_COMMITTER_DATE", "2025-01-01T00:00:00Z")
			utils.RunSucceedCommand(ctx, dir, "git", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "new HEAD")

			var repo GitRepo
			if remoteSource {
				remote, err := OpenRemoteRepo("metadata", dir, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(remote.CloneAndFetch(ctx)).To(gomega.Succeed())
				repo = remote
			} else {
				local, err := OpenLocalRepo(ctx, "metadata", dir, OpenLocalRepoOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				repo = local
			}
			metadata, err := repo.CommitMetadata(ctx, first)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(metadata.Hash).To(gomega.Equal(first))
			gomega.Expect(metadata.CommitterDate.Format(time.RFC3339)).To(gomega.Equal("2024-02-03T09:34:56Z"))
		},
		ginkgo.Entry("local", false),
		ginkgo.Entry("remote", true),
	)
})
