package git_repo_test

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/opstats"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = ginkgo.Describe("Git data cache lookup counters", func() {
	var collector *opstats.Collector
	var projectDir string

	rows := func(ctx ginkgo.SpecContext) map[string]opstats.CacheSummary {
		res := make(map[string]opstats.CacheSummary)
		for _, summary := range collector.CacheSummary(ctx) {
			res[string(summary.Operation)+"/"+string(summary.Layer)] = summary
		}
		return res
	}

	commit := func(ctx ginkgo.SpecContext, name, content string) string {
		gomega.Expect(os.MkdirAll(filepath.Join(projectDir, "app"), 0o755)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(filepath.Join(projectDir, "app", name), []byte(content), 0o644)).To(gomega.Succeed())
		utils.RunSucceedCommand(ctx, projectDir, "git", "add", ".")
		utils.RunSucceedCommand(ctx, projectDir, "git", "-c", "commit.gpgsign=false", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", name)
		return utils.GetHeadCommit(ctx, projectDir)
	}

	// openRepo returns a repository object with an empty in-memory cache, which is
	// what a new werf process starts with while the on-disk git data survives.
	openRepo := func(ctx ginkgo.SpecContext) *git_repo.Local {
		repo, err := git_repo.OpenLocalRepo(ctx, "own", projectDir, git_repo.OpenLocalRepoOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return repo
	}

	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(git_repo.Init(gitDataManager)).To(gomega.Succeed())

		collector = opstats.NewCollector()
		projectDir = ginkgo.GinkgoT().TempDir()
		utils.RunSucceedCommand(ctx, projectDir, "git", "-c", "init.defaultBranch=main", "init")
		utils.RunSucceedCommand(ctx, projectDir, "git", "config", "user.name", "Test")
		utils.RunSucceedCommand(ctx, projectDir, "git", "config", "user.email", "test@example.com")
	})

	ginkgo.It("counts one archive lookup per layer the call actually reaches", func(ctx ginkgo.SpecContext) {
		head := commit(ctx, "included.txt", "included\n")
		opts := git_repo.ArchiveOptions{
			Commit:      head,
			PathScope:   "app",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "app"}),
		}
		countedCtx := opstats.NewContext(ctx, collector)

		repo := openRepo(ctx)
		artifact, err := repo.GetOrCreateArchive(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: archive/memory": {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerMemory, Miss: 1},
			"git: archive/disk":   {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerDisk, Miss: 1},
		}))

		// A memory hit must not reach the disk layer, so the disk row stays as it was.
		_, err = repo.GetOrCreateArchive(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: archive/memory": {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 1},
			"git: archive/disk":   {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerDisk, Miss: 1},
		}))

		// A fresh process misses in memory and hits the git data kept on disk.
		_, err = openRepo(ctx).GetOrCreateArchive(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: archive/memory": {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 2},
			"git: archive/disk":   {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 1},
		}))

		gomega.Expect(os.Remove(artifact.GetFilePath())).To(gomega.Succeed())
		rebuilt, err := repo.GetOrCreateArchive(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.ReadFile(rebuilt.GetFilePath())).NotTo(gomega.BeEmpty())
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: archive/memory": {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 3},
			"git: archive/disk":   {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 2},
		}))

		gomega.Expect(os.WriteFile(strings.TrimSuffix(rebuilt.GetFilePath(), ".tar")+".meta.json", []byte("invalid"), 0o644)).To(gomega.Succeed())
		_, err = repo.GetOrCreateArchive(countedCtx, opts)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: archive/memory": {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 4},
			"git: archive/disk":   {Operation: opstats.OperationGitArchive, Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 2},
		}))
	})

	ginkgo.It("counts one patch lookup per layer the call actually reaches", func(ctx ginkgo.SpecContext) {
		fromCommit := commit(ctx, "included.txt", "included\n")
		toCommit := commit(ctx, "included.txt", "changed\n")
		opts := git_repo.PatchOptions{
			FromCommit:  fromCommit,
			ToCommit:    toCommit,
			PathScope:   "app",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "app"}),
			WithBinary:  true,
		}
		countedCtx := opstats.NewContext(ctx, collector)

		repo := openRepo(ctx)
		artifact, err := repo.GetOrCreatePatch(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = repo.GetOrCreatePatch(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = openRepo(ctx).GetOrCreatePatch(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: patch/memory": {Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 2},
			"git: patch/disk":   {Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 1},
		}))

		gomega.Expect(os.Remove(artifact.GetFilePath())).To(gomega.Succeed())
		rebuilt, err := repo.GetOrCreatePatch(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.ReadFile(rebuilt.GetFilePath())).NotTo(gomega.BeEmpty())
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: patch/memory": {Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 3},
			"git: patch/disk":   {Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 2},
		}))
	})
})
