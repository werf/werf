package image

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func TestBuildContextArchive(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Build context archive")
}

var _ = ginkgo.Describe("private context archive", func() {
	ginkgo.It("creates a Dockerfile context that survives removal of shared Git archives", func(ctx ginkgo.SpecContext) {
		root := ginkgo.GinkgoT().TempDir()
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		manager, err := gitdata.GetHostGitDataManager(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(git_repo.Init(manager)).To(gomega.Succeed())
		source := filepath.Join(root, "source")
		gomega.Expect(os.MkdirAll(source, 0o700)).To(gomega.Succeed())
		utils.RunSucceedCommand(ctx, source, "git", "init")
		gomega.Expect(os.WriteFile(filepath.Join(source, "payload.txt"), []byte("build context payload"), 0o600)).To(gomega.Succeed())
		utils.RunSucceedCommand(ctx, source, "git", "add", ".")
		utils.RunSucceedCommand(ctx, source, "git", "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "fixture")
		repo, err := git_repo.OpenLocalRepo(ctx, "fixture", source, git_repo.OpenLocalRepoOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		mgr, err := giterminism_manager.NewManager(ctx, "", source, repo, utils.GetHeadCommit(ctx, source), giterminism_manager.NewManagerOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		archive := NewBuildContextArchive(mgr, filepath.Join(root, "command"))
		gomega.Expect(archive.Create(ctx, container_backend.BuildContextArchiveCreateOptions{})).To(gomega.Succeed())
		before, err := os.ReadFile(archive.Path())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(bytes.Contains(before, []byte("build context payload"))).To(gomega.BeTrue())
		gomega.Expect(os.RemoveAll(filepath.Join(werf.GetLocalCacheDir(), "git_archives"))).To(gomega.Succeed())
		gomega.Expect(os.ReadFile(archive.Path())).To(gomega.Equal(before))
	})

	ginkgo.It("keeps each consumer readable when cached input is deleted and recreated", func(ctx ginkgo.SpecContext) {
		project := newProjectRepo(ctx, contextStreamingProjectFiles())
		first := dockerfileContextArchives(ctx, project, "one")[0]
		firstEntries := openedContextEntries(ctx, first)
		gomega.Expect(os.RemoveAll(filepath.Join(werf.GetLocalCacheDir(), "git_archives"))).To(gomega.Succeed())
		commitFiles(ctx, project, map[string]string{"blob.bin": "new content"})
		second := dockerfileContextArchives(ctx, project, "one")[0]
		gomega.Expect(os.RemoveAll(filepath.Join(werf.GetLocalCacheDir(), "git_archives"))).To(gomega.Succeed())
		gomega.Expect(openedContextEntries(ctx, first)).To(gomega.Equal(firstEntries))
		gomega.Expect(lastEntryContents(openedContextEntries(ctx, first))["blob.bin"]).NotTo(gomega.Equal("new content"))
		gomega.Expect(lastEntryContents(openedContextEntries(ctx, second))["blob.bin"]).To(gomega.Equal("new content"))
	})
})
