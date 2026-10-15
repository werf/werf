package true_git

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("submodule shallow checkout from remotes", func() {
	var (
		baseDir     string
		superRepo   string
		superGitDir string
		workTreeDir string
	)

	BeforeEach(func() {
		isolateGitConfig()
		// Shallow clones are honored over file:// only, and the remote update does not relax the
		// file transport itself — a real https remote would need nothing of the sort.
		setEnvForSpec("GIT_ALLOW_PROTOCOL", "file:http")
		baseDir = SuiteData.TestDirPath
		superRepo = filepath.Join(baseDir, "super")
		superGitDir = filepath.Join(superRepo, ".git")
		workTreeDir = filepath.Join(baseDir, "worktree")
	})

	commitFile := func(ctx context.Context, repoDir, fileName, content string) string {
		Expect(os.WriteFile(filepath.Join(repoDir, fileName), []byte(content), 0o644)).To(Succeed())
		gitSucceed(ctx, repoDir, "add", ".")
		gitSucceed(ctx, repoDir, "commit", "-m", content)
		return gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
	}

	// A remote with history behind the commit the superproject will pin, and more history after it,
	// so a clone that fetched anything beyond the pinned commit is told apart from a shallow one.
	initRemoteWithHistory := func(ctx context.Context, dir string) (first, pinned string) {
		gitInitRepo(ctx, dir)
		first = commitFile(ctx, dir, "file.txt", "first")
		pinned = commitFile(ctx, dir, "file.txt", "pinned")
		commitFile(ctx, dir, "file.txt", "tip")
		return first, pinned
	}

	// The superproject must not hold the submodule objects itself, or the update would reuse them
	// instead of going to the remote at all.
	dropSuperprojectStore := func(ctx context.Context) {
		gitSucceed(ctx, superRepo, "submodule", "deinit", "-f", "--all")
		Expect(os.RemoveAll(filepath.Join(superGitDir, "modules"))).To(Succeed())
	}

	addWorkTree := func(ctx context.Context, commit string) {
		gitSucceed(ctx, superRepo, "worktree", "add", "--detach", workTreeDir, commit)
	}

	headSHA := func(ctx context.Context, dir string) string {
		return gitSucceedTrimmed(ctx, dir, "rev-parse", "HEAD")
	}

	// `submodule add` records the remote's HEAD as the gitlink, so the remote was detached at the
	// commit to pin. Put it back on its branch: a clone sees the branch tip, not the pin, which is
	// what forces the separate depth-limited fetch of the pinned commit this suite is about.
	restoreRemoteTip := func(ctx context.Context, remote, pinned string) {
		gitSucceed(ctx, remote, "checkout", "-q", "main")
		Expect(headSHA(ctx, remote)).ToNot(Equal(pinned), "the pinned commit must not be the remote's tip, or the pin fetch would never run")
	}

	switchWorkTreeSubmodules := func(ctx context.Context, commit string) error {
		gitSucceed(ctx, workTreeDir, "checkout", "--force", "--detach", commit)
		if err := syncSubmodules(ctx, superGitDir, workTreeDir); err != nil {
			return err
		}
		return updateSubmodules(ctx, superGitDir, workTreeDir)
	}

	worktreeModuleDir := func(ctx context.Context, names ...string) string {
		worktreeGitDir, _, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
		Expect(err).ToNot(HaveOccurred())
		dir := worktreeGitDir
		for _, name := range names {
			dir = filepath.Join(dir, "modules", name)
		}
		return dir
	}

	expectFileContent := func(path, expected string) {
		content, err := os.ReadFile(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(expected))
	}

	expectCommitPresent := func(ctx context.Context, moduleDir, commit string, present bool) {
		GinkgoHelper()
		has, err := gitObjectExists(ctx, moduleDir, commit+"^{commit}")
		Expect(err).ToNot(HaveOccurred())
		Expect(has).To(Equal(present), "commit %s in %s", commit, moduleDir)
	}

	It("clones a missing submodule shallow at its pinned commit and pins the commit with a ref", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		first, pinned := initRemoteWithHistory(ctx, subRemote)
		gitSucceed(ctx, subRemote, "checkout", "-q", pinned)

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
		Expect(gitSucceedTrimmed(ctx, filepath.Join(superRepo, "sub"), "rev-parse", "HEAD")).To(Equal(pinned))
		restoreRemoteTip(ctx, subRemote, pinned)
		dropSuperprojectStore(ctx)

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "pinned")
		moduleDir := worktreeModuleDir(ctx, "sub")
		Expect(filepath.Join(moduleDir, "shallow")).To(BeARegularFile())
		expectCommitPresent(ctx, moduleDir, pinned, true)
		expectCommitPresent(ctx, moduleDir, first, false)
		Expect(gitSucceedTrimmed(ctx, moduleDir, "rev-parse", "--verify", "refs/werf/commits/"+pinned)).To(Equal(pinned))
	})

	It("clones nested submodules shallow too", func(ctx SpecContext) {
		leafRemote := filepath.Join(baseDir, "leaf-remote")
		leafFirst, leafPinned := initRemoteWithHistory(ctx, leafRemote)
		gitSucceed(ctx, leafRemote, "checkout", "-q", leafPinned)

		midRemote := filepath.Join(baseDir, "mid-remote")
		gitInitRepo(ctx, midRemote)
		commitFile(ctx, midRemote, "mid.txt", "first")
		gitAddSubmoduleSucceed(ctx, midRemote, "file://"+leafRemote, "leaf")
		gitSucceed(ctx, midRemote, "commit", "-m", "add leaf")
		restoreRemoteTip(ctx, leafRemote, leafPinned)
		midPinned := headSHA(ctx, midRemote)
		commitFile(ctx, midRemote, "mid.txt", "tip")
		gitSucceed(ctx, midRemote, "checkout", "-q", midPinned)

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+midRemote, "mid")
		gitSucceed(ctx, superRepo, "commit", "-m", "add mid")
		restoreRemoteTip(ctx, midRemote, midPinned)
		dropSuperprojectStore(ctx)

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "mid", "mid.txt"), "first")
		expectFileContent(filepath.Join(workTreeDir, "mid", "leaf", "file.txt"), "pinned")
		expectSubmoduleFile(superGitDir, workTreeDir, "file.txt", "pinned", "mid", "leaf")
		Expect(filepath.Join(worktreeModuleDir(ctx, "mid"), "shallow")).To(BeARegularFile())
		leafModuleDir := worktreeModuleDir(ctx, "mid", "leaf")
		Expect(filepath.Join(leafModuleDir, "shallow")).To(BeARegularFile())
		expectCommitPresent(ctx, leafModuleDir, leafFirst, false)
		Expect(gitSucceedTrimmed(ctx, leafModuleDir, "rev-parse", "--verify", "refs/werf/commits/"+leafPinned)).To(Equal(leafPinned))
	})

	// In a shallow clone nothing but a detached HEAD holds the previous commit, so switching the
	// gitlink back would fetch it again. The pinning ref must make that switch work offline.
	It("switches back to an earlier pinned commit without the remote", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepo(ctx, subRemote)
		commitFile(ctx, subRemote, "file.txt", "first")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
		firstSuperSHA := headSHA(ctx, superRepo)

		commitFile(ctx, subRemote, "file.txt", "second")
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--remote", "sub")
		gitSucceed(ctx, superRepo, "commit", "-am", "bump submodule")
		secondSuperSHA := headSHA(ctx, superRepo)
		dropSuperprojectStore(ctx)

		addWorkTree(ctx, firstSuperSHA)
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "first")

		Expect(switchWorkTreeSubmodules(ctx, secondSuperSHA)).To(Succeed())
		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
		Expect(filepath.Join(worktreeModuleDir(ctx, "sub"), "shallow")).To(BeARegularFile(), "the moved gitlink must keep the clone shallow")

		Expect(os.RemoveAll(subRemote)).To(Succeed())

		Expect(switchWorkTreeSubmodules(ctx, firstSuperSHA)).To(Succeed())
		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "first")
	})

	// A service worktree populated before shallow checkouts existed holds full clones. Fetching a
	// moved gitlink into one with --depth would turn it shallow and lose the history it already has.
	It("keeps an existing full clone full when its gitlink moves", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepo(ctx, subRemote)
		first := commitFile(ctx, subRemote, "file.txt", "first")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
		dropSuperprojectStore(ctx)

		addWorkTree(ctx, headSHA(ctx, superRepo))
		gitUpdateSubmodulesSucceed(ctx, workTreeDir, "--init")
		moduleDir := worktreeModuleDir(ctx, "sub")
		Expect(filepath.Join(moduleDir, "shallow")).ToNot(BeAnExistingFile())

		// Only now does the remote move on, so the new gitlink has to be fetched into the full clone.
		second := commitFile(ctx, subRemote, "file.txt", "second")
		expectCommitPresent(ctx, moduleDir, second, false)
		gitSucceed(ctx, superRepo, "update-index", "--cacheinfo", "160000,"+second+",sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "bump submodule")

		Expect(switchWorkTreeSubmodules(ctx, headSHA(ctx, superRepo))).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
		Expect(filepath.Join(moduleDir, "shallow")).ToNot(BeAnExistingFile())
		expectCommitPresent(ctx, moduleDir, first, true)
	})

	DescribeTable("clones in full when shallow is ruled out",
		func(ctx SpecContext, ruleOut func(ctx context.Context)) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			first, pinned := initRemoteWithHistory(ctx, subRemote)
			gitSucceed(ctx, subRemote, "checkout", "-q", pinned)

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
			ruleOut(ctx)
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
			restoreRemoteTip(ctx, subRemote, pinned)
			dropSuperprojectStore(ctx)

			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "pinned")
			moduleDir := worktreeModuleDir(ctx, "sub")
			Expect(filepath.Join(moduleDir, "shallow")).ToNot(BeAnExistingFile())
			expectCommitPresent(ctx, moduleDir, first, true)
		},
		Entry("by shallow=false in .gitmodules", func(ctx context.Context) {
			gitSucceed(ctx, superRepo, "config", "-f", ".gitmodules", "submodule.sub.shallow", "false")
			gitSucceed(ctx, superRepo, "add", ".gitmodules")
		}),
		Entry("by "+disableShallowSubmodulesEnvVar, func(ctx context.Context) {
			setEnvForSpec(disableShallowSubmodulesEnvVar, "1")
		}),
	)

	// The dumb HTTP transport cannot serve a shallow clone at all. Such a submodule must still be
	// checked out — in full — while its siblings keep the shallow clone.
	It("falls back to a full clone for a submodule whose remote refuses shallow, keeping its sibling shallow", func(ctx SpecContext) {
		dumbRemote := filepath.Join(baseDir, "dumb-remote")
		dumbFirst, dumbPinned := initRemoteWithHistory(ctx, dumbRemote)
		dumbBare := filepath.Join(baseDir, "served", "dumb.git")
		gitSucceed(ctx, baseDir, "clone", "--bare", dumbRemote, dumbBare)
		gitSucceed(ctx, dumbBare, "update-server-info")
		server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join(baseDir, "served"))))
		DeferCleanup(server.Close)

		subRemote := filepath.Join(baseDir, "sub-remote")
		_, subPinned := initRemoteWithHistory(ctx, subRemote)
		gitSucceed(ctx, subRemote, "checkout", "-q", subPinned)

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "a-sub")
		gitAddSubmoduleSucceed(ctx, superRepo, server.URL+"/dumb.git", "b-dumb")
		gitSucceed(ctx, filepath.Join(superRepo, "b-dumb"), "checkout", "-q", dumbPinned)
		gitSucceed(ctx, superRepo, "add", "b-dumb")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodules")
		restoreRemoteTip(ctx, subRemote, subPinned)
		dropSuperprojectStore(ctx)

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "a-sub", "file.txt"), "pinned")
		expectFileContent(filepath.Join(workTreeDir, "b-dumb", "file.txt"), "pinned")
		Expect(filepath.Join(worktreeModuleDir(ctx, "a-sub"), "shallow")).To(BeARegularFile())
		dumbModuleDir := worktreeModuleDir(ctx, "b-dumb")
		Expect(filepath.Join(dumbModuleDir, "shallow")).ToNot(BeAnExistingFile())
		expectCommitPresent(ctx, dumbModuleDir, dumbFirst, true)
		Expect(gitSucceedTrimmed(ctx, dumbModuleDir, "rev-parse", "--verify", "refs/werf/commits/"+dumbPinned)).To(Equal(dumbPinned))
	})

	// A shallow clone the remote can no longer complete must be replaced, not fetched into: a plain
	// fetch keeps it shallow, so the full retry would never be full.
	It("replaces a shallow clone that lacks the moved gitlink with a full one when the shallow fetch fails", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepo(ctx, subRemote)
		first := commitFile(ctx, subRemote, "file.txt", "first")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
		dropSuperprojectStore(ctx)

		// The shallow clone happens while the remote's tip IS the pinned commit, so the commit the
		// gitlink moves to below is not in the clone yet and has to be fetched.
		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		moduleDir := worktreeModuleDir(ctx, "sub")
		Expect(filepath.Join(moduleDir, "shallow")).To(BeARegularFile())

		second := commitFile(ctx, subRemote, "file.txt", "second")
		expectCommitPresent(ctx, moduleDir, second, false)

		// The remote moves to a dumb HTTP server, which can serve a full clone but no shallow fetch.
		dumbBare := filepath.Join(baseDir, "served", "sub.git")
		gitSucceed(ctx, baseDir, "clone", "--bare", subRemote, dumbBare)
		gitSucceed(ctx, dumbBare, "update-server-info")
		server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join(baseDir, "served"))))
		DeferCleanup(server.Close)
		gitSucceed(ctx, superRepo, "config", "-f", ".gitmodules", "submodule.sub.url", server.URL+"/sub.git")
		gitSucceed(ctx, superRepo, "update-index", "--cacheinfo", "160000,"+second+",sub")
		gitSucceed(ctx, superRepo, "commit", "-am", "move submodule remote and bump gitlink")
		Expect(os.RemoveAll(subRemote)).To(Succeed())

		Expect(switchWorkTreeSubmodules(ctx, headSHA(ctx, superRepo))).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
		Expect(filepath.Join(moduleDir, "shallow")).ToNot(BeAnExistingFile())
		expectCommitPresent(ctx, moduleDir, first, true)
		expectCommitPresent(ctx, moduleDir, second, true)
	})

	It("fails when the pinned commit is not on the remote instead of checking out another one", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepo(ctx, subRemote)
		commitFile(ctx, subRemote, "file.txt", "kept")
		commitFile(ctx, subRemote, "file.txt", "lost")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
		dropSuperprojectStore(ctx)

		gitSucceed(ctx, subRemote, "reset", "-q", "--hard", "HEAD~1")
		gitSucceed(ctx, subRemote, "reflog", "expire", "--expire=now", "--all")
		gitSucceed(ctx, subRemote, "gc", "-q", "--prune=now")

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		err := updateSubmodules(ctx, superGitDir, workTreeDir)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("update submodule sub:"))
		Expect(filepath.Join(workTreeDir, "sub", "file.txt")).ToNot(BeAnExistingFile())
	})
})
