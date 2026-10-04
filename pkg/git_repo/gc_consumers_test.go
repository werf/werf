package git_repo_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/git_repo/gitdata"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

// These specs drive the production gitdata.GitDataManager, so the cache layout
// and metadata they assert against are the ones a real build and a real
// `werf host cleanup` share.
var _ = Describe("Git cache consumers under GC eviction", func() {
	var manager *gitdata.GitDataManager
	var sourceDir string

	gitInSource := func(ctx context.Context, args ...string) {
		utils.RunSucceedCommand(ctx, sourceDir, "git", append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.com"}, args...)...)
	}

	commitFile := func(ctx context.Context, name, content string) string {
		utils.WriteFile(filepath.Join(sourceDir, name), []byte(content))
		gitInSource(ctx, "add", "--", name)
		gitInSource(ctx, "commit", "-m", "commit "+name+" "+content)
		return utils.GetHeadCommit(ctx, sourceDir)
	}

	BeforeEach(func(ctx SpecContext) {
		tmpDir := GinkgoT().TempDir()
		Expect(werf.Init(tmpDir, GinkgoT().TempDir())).To(Succeed())
		Expect(true_git.Init(ctx, true_git.Options{})).To(Succeed())

		var err error
		manager, err = gitdata.GetHostGitDataManager(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(git_repo.Init(manager)).To(Succeed())

		sourceDir = filepath.Join(tmpDir, "source")
		Expect(os.MkdirAll(sourceDir, 0o755)).To(Succeed())
		utils.RunSucceedCommand(ctx, sourceDir, "git", "-c", "init.defaultBranch=main", "init")
		// A real server serving a shallow commit mapping must allow by-SHA
		// fetches; without it werf would downgrade the mapping to a full mirror.
		utils.RunSucceedCommand(ctx, sourceDir, "git", "config", "uploadpack.allowReachableSHA1InWant", "true")
	})

	Describe("archives and patches", func() {
		It("rebuilds an archive evicted from disk after a warm memory lookup", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo, err := git_repo.OpenLocalRepo(ctx, "own", sourceDir, git_repo.OpenLocalRepoOptions{})
			Expect(err).NotTo(HaveOccurred())

			opts := git_repo.ArchiveOptions{
				Commit:      commit,
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			}

			first, err := repo.GetOrCreateArchive(ctx, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.ReadFile(first.GetFilePath())).NotTo(BeEmpty())

			// GC evicts the artifact while the repo object keeps it in memory.
			Expect(os.Remove(first.GetFilePath())).To(Succeed())

			second, err := repo.GetOrCreateArchive(ctx, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.ReadFile(second.GetFilePath())).NotTo(BeEmpty())
		})

		It("rebuilds a patch evicted from disk after a warm memory lookup", func(ctx SpecContext) {
			fromCommit := commitFile(ctx, "data.txt", "v1")
			toCommit := commitFile(ctx, "data.txt", "v2")

			repo, err := git_repo.OpenLocalRepo(ctx, "own", sourceDir, git_repo.OpenLocalRepoOptions{})
			Expect(err).NotTo(HaveOccurred())

			opts := git_repo.PatchOptions{
				FromCommit:  fromCommit,
				ToCommit:    toCommit,
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			}

			first, err := repo.GetOrCreatePatch(ctx, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.ReadFile(first.GetFilePath())).NotTo(BeEmpty())
			Expect(first.GetPaths()).To(ConsistOf("data.txt"))

			Expect(os.Remove(first.GetFilePath())).To(Succeed())

			second, err := repo.GetOrCreatePatch(ctx, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.ReadFile(second.GetFilePath())).NotTo(BeEmpty())
			Expect(second.GetPaths()).To(ConsistOf("data.txt"))
		})
	})
})
