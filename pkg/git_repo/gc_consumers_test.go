package git_repo_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/git_repo/gitdata"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

// These specs drive the production gitdata.GitDataManager (same cache layout,
// same "git_data_manager" GC lock) so that the GC exclusion they assert is the
// one a real `werf host cleanup` takes.
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

	openRemoteURL := func(url, branch, tag, commit string) *git_repo.Remote {
		repo, err := git_repo.OpenRemoteRepo("test-mapping", url, nil)
		Expect(err).NotTo(HaveOccurred())
		repo.Branch = branch
		repo.Tag = tag
		repo.Commit = commit
		return repo
	}

	openRemote := func(branch, tag, commit string) *git_repo.Remote {
		repo, err := git_repo.OpenRemoteRepo("test-mapping", sourceDir, nil)
		Expect(err).NotTo(HaveOccurred())
		repo.Branch = branch
		repo.Tag = tag
		repo.Commit = commit
		return repo
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

	Describe("remote mirrors", func() {
		It("restores an evicted full mirror and keeps serving the originally resolved commit", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			// The branch moves on, then GC takes the mirror the opened repo uses.
			commitFile(ctx, "data.txt", "v2")
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
		})

		It("restores an evicted shallow mirror and keeps serving the mapped commit", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("", "", commit)
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.GetClonePath()).To(ContainSubstring("git_mirrors"))
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			commitFile(ctx, "data.txt", "v2")
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
		})

		It("fails with a bounded error when the evicted commit is gone from the origin too", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())
			Expect(os.RemoveAll(sourceDir)).To(Succeed())

			_, err := repo.ReadCommitFile(ctx, commit, "data.txt")
			Expect(err).To(HaveOccurred())
		})

		It("refetches a commit no longer advertised by any branch", func(ctx SpecContext) {
			commitFile(ctx, "data.txt", "base")
			gitInSource(ctx, "checkout", "-b", "side")
			commit := commitFile(ctx, "data.txt", "v1")
			gitInSource(ctx, "checkout", "main")

			// A file:// origin is served by git-upload-pack like a real remote,
			// without the object copying a local path clone does.
			utils.RunSucceedCommand(ctx, sourceDir, "git", "config", "uploadpack.allowAnySHA1InWant", "true")
			repo := openRemoteURL("file://"+sourceDir, "", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			// The branch holding the commit is deleted, then GC takes the mirror:
			// fetching all refs no longer brings the commit back, only asking the
			// origin for the SHA itself does.
			gitInSource(ctx, "branch", "-D", "side")
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
		})

		It("fails with a bounded error when the origin refuses the commit", func(ctx SpecContext) {
			commitFile(ctx, "data.txt", "base")
			gitInSource(ctx, "checkout", "-b", "side")
			commit := commitFile(ctx, "data.txt", "v1")
			gitInSource(ctx, "checkout", "main")

			repo := openRemoteURL("file://"+sourceDir, "", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			// The origin itself drops the commit, so no fetch can bring it back.
			gitInSource(ctx, "branch", "-D", "side")
			gitInSource(ctx, "reflog", "expire", "--expire=now", "--all")
			gitInSource(ctx, "gc", "--prune=now", "--quiet")
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			_, err := repo.ReadCommitFile(ctx, commit, "data.txt")
			Expect(err).To(MatchError(ContainSubstring("commit " + commit + " is not available in origin")))
		})

		It("resolves the branch commit after the mirror was evicted", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			// Nothing resolved this branch yet: GetLatestCommitInfo asks for it
			// per stage, and by then GC may already have taken the mirror.
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			Expect(repo.LatestBranchCommit(ctx, "main")).To(Equal(commit))
		})

		It("keeps serving the branch commit resolved before the mirror was evicted", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.LatestBranchCommit(ctx, "main")).To(Equal(commit))

			// The origin advances and GC takes the mirror. Re-resolving the
			// branch would pin later stages of this build to a different commit
			// than the ones already built.
			commitFile(ctx, "data.txt", "v2")
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			Expect(repo.LatestBranchCommit(ctx, "main")).To(Equal(commit))
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
		})

		It("keeps serving the tag commit resolved before the mirror was evicted", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")
			gitInSource(ctx, "tag", "v1")

			repo := openRemote("", "v1", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.TagCommit(ctx, "v1")).To(Equal(commit))

			movedCommit := commitFile(ctx, "data.txt", "v2")
			gitInSource(ctx, "tag", "-f", "v1", movedCommit)
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())

			Expect(repo.TagCommit(ctx, "v1")).To(Equal(commit))
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
		})

		It("re-resolves the branch commit after an explicit fetch", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.LatestBranchCommit(ctx, "main")).To(Equal(commit))

			advanced := commitFile(ctx, "data.txt", "v2")
			Expect(repo.FetchOrigin(ctx, git_repo.FetchOptions{})).To(Succeed())

			Expect(repo.LatestBranchCommit(ctx, "main")).To(Equal(advanced))
		})

		It("keeps reading through a handle whose mirror was recreated by another process", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			// Another werf process loses the mirror to GC and clones it again:
			// the path exists, but every file behind the cached handle is new.
			Expect(os.RemoveAll(repo.GetClonePath())).To(Succeed())
			peer := openRemote("main", "", "")
			Expect(peer.CloneAndFetch(ctx)).To(Succeed())

			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
		})

		It("refreshes the mirror last access timestamp on a read", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			lastAccessPath := filepath.Join(repo.GetClonePath(), "last_access_at")
			backdated := time.Now().Add(-7 * 24 * time.Hour)
			Expect(timestamps.WriteTimestampFile(lastAccessPath, backdated)).To(Succeed())

			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			refreshed, err := timestamps.ReadTimestampFile(lastAccessPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(refreshed).To(BeTemporally(">", backdated.Add(time.Hour)))
		})

		It("keeps the original go-git object-not-found retry working", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))

			// A mirror that is present but missing the object files: the handle
			// retry path must surface the error instead of looping.
			Expect(os.RemoveAll(filepath.Join(repo.GetClonePath(), "objects"))).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(repo.GetClonePath(), "objects"), 0o755)).To(Succeed())

			_, err := repo.ReadCommitFile(ctx, commit, "data.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GC exclusion", func() {
		It("runs repo operations under a GC lock the caller already holds", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			// Dockerfile build context archiving takes the shared GC lock around
			// git operations that take it again.
			outerLock, err := manager.LockGC(ctx, true)
			Expect(err).NotTo(HaveOccurred())
			defer func() { Expect(werf.HostLocker().ReleaseLock(outerLock)).To(Succeed()) }()

			done := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(done)
				Expect(repo.ReadCommitFile(ctx, commit, "data.txt")).To(Equal([]byte("v1")))
				Expect(repo.LatestBranchCommit(ctx, "main")).To(Equal(commit))
			}()

			Eventually(done, 30*time.Second).Should(BeClosed())
		})

		It("keeps GC out while a repo handle is in use and lets it in on release", func(ctx SpecContext) {
			commit := commitFile(ctx, "data.txt", "v1")

			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			insideHandle := make(chan struct{})
			releaseHandle := make(chan struct{})
			handleDone := make(chan error, 1)

			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseHandle) }) }
			DeferCleanup(release)

			go func() {
				defer GinkgoRecover()
				handleDone <- git_repo.WithRepoHandleForTest(ctx, repo, commit, func() error {
					close(insideHandle)
					<-releaseHandle
					return nil
				})
			}()

			Eventually(insideHandle).Should(BeClosed())

			gcEntered := make(chan struct{})
			gcDone := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(gcDone)
				lock, err := manager.LockGC(ctx, false)
				Expect(err).NotTo(HaveOccurred())
				close(gcEntered)
				Expect(werf.HostLocker().ReleaseLock(lock)).To(Succeed())
			}()

			Consistently(gcEntered, 500*time.Millisecond, 50*time.Millisecond).ShouldNot(BeClosed())

			release()
			Eventually(gcEntered, 30*time.Second).Should(BeClosed())
			Eventually(gcDone, 30*time.Second).Should(BeClosed())
			Expect(<-handleDone).To(Succeed())
		})
	})
})
