package true_git

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("include.path forwarding", func() {
	var (
		baseDir string
		repoDir string
		gitDir  string
	)

	BeforeEach(func() {
		baseDir = SuiteData.TestDirPath
		repoDir = filepath.Join(baseDir, "repo")
		gitDir = filepath.Join(repoDir, ".git")
	})

	// A URL rewrite reachable only through include.path: git applies it if and only if werf forwards
	// the include, so the marker showing up is what proves the forwarding.
	writeIncludedRewrite := func(ctx context.Context) {
		stubPath := filepath.Join(baseDir, "ext.conf")
		stub := "[url \"werf-test-marker://rewritten/\"]\n\tinsteadOf = https://\n"
		Expect(os.WriteFile(stubPath, []byte(stub), 0o644)).To(Succeed())
		gitSucceed(ctx, repoDir, "config", "include.path", stubPath)
	}

	It("resolves every include.path of the repository config", func(ctx SpecContext) {
		gitInitRepo(ctx, repoDir)

		opts, err := getIncludePathOptions(ctx, gitDir)
		Expect(err).ToNot(HaveOccurred())
		Expect(opts).To(BeEmpty())

		gitSucceed(ctx, repoDir, "config", "--add", "include.path", "/abs/ext.conf")
		gitSucceed(ctx, repoDir, "config", "--add", "include.path", "rel/ext.conf")
		gitSucceed(ctx, repoDir, "config", "--add", "include.path", "~/tilde-ext.conf")

		opts, err = getIncludePathOptions(ctx, gitDir)
		Expect(err).ToNot(HaveOccurred())
		Expect(opts).To(HaveLen(6))
		Expect(opts[0]).To(Equal("-c"))
		Expect(opts[1]).To(Equal("include.path=/abs/ext.conf"))
		Expect(opts[2]).To(Equal("-c"))
		Expect(opts[3]).To(HavePrefix("include.path="))
		relResolved := strings.TrimPrefix(opts[3], "include.path=")
		Expect(filepath.IsAbs(relResolved)).To(BeTrue())
		Expect(relResolved).To(HaveSuffix(filepath.Join(".git", "rel", "ext.conf")))
		Expect(opts[4]).To(Equal("-c"))
		Expect(opts[5]).To(Equal("include.path=~/tilde-ext.conf"))
	})

	It("forwards the include to the submodule update", func(ctx SpecContext) {
		gitInitRepo(ctx, repoDir)
		headSHA := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")

		gitmodules := "[submodule \"sub\"]\n\tpath = sub\n\turl = https://werf-test-nonexistent.invalid/sub.git\n"
		Expect(os.WriteFile(filepath.Join(repoDir, ".gitmodules"), []byte(gitmodules), 0o644)).To(Succeed())
		gitSucceed(ctx, repoDir, "update-index", "--add", "--cacheinfo", "160000,"+headSHA+",sub")
		gitSucceed(ctx, repoDir, "add", ".gitmodules")
		gitSucceed(ctx, repoDir, "commit", "-m", "add submodule")

		writeIncludedRewrite(ctx)

		err := updateSubmodules(ctx, gitDir, repoDir)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("werf-test-marker"))
	})

	It("leaves a work tree switch without submodules unaffected", func(ctx SpecContext) {
		gitInitRepo(ctx, repoDir)
		headSHA := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
		writeIncludedRewrite(ctx)

		Expect(switchWorkTree(ctx, gitDir, filepath.Join(baseDir, "worktree"), headSHA, false)).To(Succeed())
	})
})

var _ = Describe("submoduleNameUnsafe", func() {
	DescribeTable("rejects a name that cannot be carried by -c submodule.<name>.url",
		func(name string, expectedUnsafe bool) {
			Expect(submoduleNameUnsafe(name)).To(Equal(expectedUnsafe))
		},
		Entry("equals sign ends the config key early", "a=b", true),
		Entry("newline splits the config key", "a\nb", true),
		Entry("slash moves the store path", "a/b", true),
		Entry("backslash moves the store path on windows", "a\\b", true),
		Entry("single dot resolves to the store parent", ".", true),
		Entry("double dot escapes the store", "..", true),
		Entry("plain name", "sub", false),
		Entry("dash, underscore and dot inside are fine", "my-sub_2.0", false),
		Entry("a dot-prefixed name is not a dot component", ".sub", false),
		Entry("a name ending in a dot is not a dot component", "sub.", false),
		Entry("three dots are not a dot component", "...", false),
	)
})

var _ = Describe("submodule local object store reuse", func() {
	var (
		baseDir     string
		superRepo   string
		superGitDir string
		workTreeDir string
	)

	BeforeEach(func() {
		isolateGitConfig()
		baseDir = SuiteData.TestDirPath
		superRepo = filepath.Join(baseDir, "super")
		superGitDir = filepath.Join(superRepo, ".git")
		workTreeDir = filepath.Join(baseDir, "worktree")
	})

	// The test "remotes" here are local paths, which the plain remote update reaches only with the
	// file transport allowed — a real https remote would need nothing of the sort.
	allowFileTransport := func() {
		setEnvForSpec("GIT_ALLOW_PROTOCOL", "file")
	}

	addWorkTree := func(ctx context.Context, commit string) {
		gitSucceed(ctx, superRepo, "worktree", "add", "--detach", workTreeDir, commit)
	}

	headSHA := func(ctx context.Context, dir string) string {
		return gitSucceedTrimmed(ctx, dir, "rev-parse", "HEAD")
	}

	expectFileContent := func(path, expected string) {
		content, err := os.ReadFile(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(expected))
	}

	It("checks a submodule out from the local store when its remote is gone", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

		// `submodule add` already populated superRepo/.git/modules/sub; drop the remote to prove the
		// checkout into a fresh worktree reuses those local objects and touches no network.
		Expect(os.RemoveAll(subRemote)).To(Succeed())

		addWorkTree(ctx, headSHA(ctx, superRepo))

		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
	})

	It("checks a nested submodule out from the local store when every remote is gone", func(ctx SpecContext) {
		leafRemote := filepath.Join(baseDir, "leaf-remote")
		gitInitRepoWithFile(ctx, leafRemote, "leaf.txt", "LEAF")

		midRemote := filepath.Join(baseDir, "mid-remote")
		gitInitRepo(ctx, midRemote)
		gitAddSubmoduleSucceed(ctx, midRemote, leafRemote, "leaf")
		gitSucceed(ctx, midRemote, "commit", "-m", "add leaf")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, midRemote, "mid")
		// `submodule add` populates only the top module store; recurse once to fill mid/modules/leaf.
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--init", "--recursive")
		gitSucceed(ctx, superRepo, "commit", "-m", "add mid")

		Expect(os.RemoveAll(leafRemote)).To(Succeed())
		Expect(os.RemoveAll(midRemote)).To(Succeed())

		addWorkTree(ctx, headSHA(ctx, superRepo))

		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "mid", "leaf", "leaf.txt"), "LEAF")
	})

	// An already-populated submodule is not re-cloned, so it is fetched from the URL recorded in the
	// service worktree's own submodule git dir. Moving the gitlink must therefore still resolve from
	// the superproject store rather than from the remote.
	It("serves a moved gitlink from the local store in an already warm work tree", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepoWithFile(ctx, subRemote, "file.txt", "first")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

		// Warm the service worktree on the first gitlink, exactly as an earlier werf run would — sync
		// included, since it rewrites the submodule remote from .gitmodules on every run.
		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		// Move the gitlink, and let the superproject store learn the new commit the way CI would.
		Expect(os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("second"), 0o644)).To(Succeed())
		gitSucceed(ctx, subRemote, "commit", "-am", "second")
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--remote", "sub")
		gitSucceed(ctx, superRepo, "commit", "-am", "bump submodule")
		secondSHA := headSHA(ctx, superRepo)
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--init")

		// Only now drop the remote: the moved commit exists locally, so no fetch may be needed.
		Expect(os.RemoveAll(subRemote)).To(Succeed())

		gitSucceed(ctx, workTreeDir, "checkout", "--force", "--detach", secondSHA)
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
	})

	// The fetch redirect must reach nested levels too: their git dirs are per-worktree just like the
	// top-level one, so a moved nested gitlink is fetched, not re-cloned.
	It("serves a moved nested gitlink from the local store in an already warm work tree", func(ctx SpecContext) {
		leafRemote := filepath.Join(baseDir, "leaf-remote")
		gitInitRepoWithFile(ctx, leafRemote, "leaf.txt", "first")

		midRemote := filepath.Join(baseDir, "mid-remote")
		gitInitRepo(ctx, midRemote)
		gitAddSubmoduleSucceed(ctx, midRemote, leafRemote, "leaf")
		gitSucceed(ctx, midRemote, "commit", "-m", "add leaf")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, midRemote, "mid")
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--init", "--recursive")
		gitSucceed(ctx, superRepo, "commit", "-m", "add mid")

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		// Move the NESTED gitlink and let the superproject stores learn the new commits.
		Expect(os.WriteFile(filepath.Join(leafRemote, "leaf.txt"), []byte("second"), 0o644)).To(Succeed())
		gitSucceed(ctx, leafRemote, "commit", "-am", "second")
		gitUpdateSubmodulesSucceed(ctx, midRemote, "--remote", "leaf")
		gitSucceed(ctx, midRemote, "commit", "-am", "bump leaf")
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--remote", "mid")
		gitSucceed(ctx, superRepo, "commit", "-am", "bump mid")
		secondSHA := headSHA(ctx, superRepo)
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--init", "--recursive")

		Expect(os.RemoveAll(leafRemote)).To(Succeed())
		Expect(os.RemoveAll(midRemote)).To(Succeed())

		gitSucceed(ctx, workTreeDir, "checkout", "--force", "--detach", secondSHA)
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "mid", "leaf", "leaf.txt"), "second")
	})

	// GitLab CI clones submodules shallow for GIT_SUBMODULE_DEPTH. A shallow store is not a partial
	// one — it has every object of the commits it holds — so it must stay eligible for reuse.
	It("reuses a shallow submodule store", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepo(ctx, subRemote)
		for _, marker := range []string{"first", "second"} {
			Expect(os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte(marker), 0o644)).To(Succeed())
			gitSucceed(ctx, subRemote, "add", ".")
			gitSucceed(ctx, subRemote, "commit", "-m", marker)
		}

		gitInitRepo(ctx, superRepo)
		// A depth-limited clone is only honored over file://, not over a plain local path.
		gitAddSubmoduleSucceed(ctx, superRepo, "file://"+subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
		gitSucceed(ctx, superRepo, "submodule", "deinit", "-f", "sub")
		Expect(os.RemoveAll(filepath.Join(superGitDir, "modules", "sub"))).To(Succeed())
		gitUpdateSubmodulesSucceed(ctx, superRepo, "--init", "--depth=1")
		Expect(filepath.Join(superGitDir, "modules", "sub", "shallow")).To(BeARegularFile())

		Expect(os.RemoveAll(subRemote)).To(Succeed())

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
	})

	// GIT_SUBMODULE_FORCE_HTTPS makes GitLab CI install its own broad insteadOf prefix rewrite. Ours
	// is keyed on the full submodule URL, and git resolves insteadOf by longest match, so reuse wins.
	It("reuses the local store despite a broader ambient insteadOf rule", func(ctx SpecContext) {
		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

		// Redirect the remote's parent prefix at a path that does not exist, standing in for the CI's
		// token rewrite: reuse must not depend on where that rule points.
		gitSucceed(ctx, superRepo, "config", "url."+filepath.Join(baseDir, "tokenized")+"/.insteadOf", filepath.Dir(subRemote)+"/")
		Expect(os.RemoveAll(subRemote)).To(Succeed())

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
	})

	// Reuse rests on git behavior that is not promised, so a store that passes every check and still
	// cannot serve the checkout must cost a slower build, not a failed one.
	It("falls back to the remotes when the local store cannot serve the checkout", func(ctx SpecContext) {
		// The retry deliberately drops the transport allowance that only the local store needs.
		allowFileTransport()

		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

		// Drop the blob from the store while leaving the pinned commit reachable: the probe still
		// succeeds, and nothing marks the store as partial, so only the checkout can discover this.
		subSHA := strings.Fields(gitSucceed(ctx, superRepo, "ls-tree", "HEAD", "--", "sub"))[2]
		storeDir := filepath.Join(superGitDir, "modules", "sub")
		blobSHA := gitSucceedTrimmed(ctx, storeDir, "rev-parse", subSHA+":file.txt")
		Expect(os.Remove(filepath.Join(storeDir, "objects", blobSHA[:2], blobSHA[2:]))).To(Succeed())

		addWorkTree(ctx, headSHA(ctx, superRepo))
		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		// Pins that the broken store is still reuse-eligible. Without this the spec would keep passing
		// through the ordinary remote path if the store ever stopped qualifying, and would no longer
		// exercise the retry it is named for.
		overrides, blockers, err := submoduleLocalURLOverrides(ctx, workTreeDir)
		Expect(err).ToNot(HaveOccurred())
		Expect(blockers).To(BeEmpty(), "a missing blob is undetectable up front, so nothing may block reuse")
		Expect(overrides).ToNot(BeEmpty())

		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")

		// The submodule ends up served by the retry, so it records the real remote, not the store.
		worktreeGitDir, _, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
		Expect(err).ToNot(HaveOccurred())
		servedFrom := gitSucceedTrimmed(ctx, filepath.Join(worktreeGitDir, "modules", "sub"), "config", "--get", "remote.origin.url")
		Expect(servedFrom).To(Equal(subRemote))
	})

	// Determining whether reuse applies is itself an optimization: when the probe cannot even run,
	// the build must get slower, not fail.
	It("falls back to the remotes when the local store cannot be inspected at all", func(ctx SpecContext) {
		allowFileTransport()

		subRemote := filepath.Join(baseDir, "sub-remote")
		gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

		gitInitRepo(ctx, superRepo)
		gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
		gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

		addWorkTree(ctx, headSHA(ctx, superRepo))

		// An unreadable config makes every probe of this store fail outright, unlike the blockers,
		// which are stores werf understands and rejects. A linked worktree keeps its own submodule git
		// dirs, so the plain remote update never touches this one.
		storeConfig := filepath.Join(superGitDir, "modules", "sub", "config")
		Expect(storeConfig).To(BeARegularFile())
		Expect(os.WriteFile(storeConfig, []byte("[this is not valid git config\n"), 0o644)).To(Succeed())

		_, _, err := submoduleLocalURLOverrides(ctx, workTreeDir)
		Expect(err).To(HaveOccurred(), "the spec is only meaningful while the probe really fails")

		Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
		Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

		expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
	})

	Describe("blockers that suppress reuse entirely", func() {
		// Partial reuse is unsafe, so every one of these must leave NO override behind: the plain
		// remote update werf has always done is the only safe fallback.
		expectReuseBlocked := func(ctx context.Context, dir, blockerSubstring string) {
			overrides, blockers, err := submoduleLocalURLOverrides(ctx, dir)
			Expect(err).ToNot(HaveOccurred())
			Expect(overrides).To(BeEmpty(), "a blocked submodule must suppress ALL overrides")
			Expect(strings.Join(blockers, "\n")).To(ContainSubstring(blockerSubstring))
		}

		// A submodule name reused at another nesting level cannot be expressed as a process-global
		// submodule.<name>.url, so no override may be emitted: an override for the top-level copy
		// would otherwise hijack the nested clone and fail the whole update.
		It("blocks on a submodule name used at two nesting levels", func(ctx SpecContext) {
			newLib := func(marker string) string {
				dir := filepath.Join(baseDir, "lib-"+marker)
				gitInitRepoWithFile(ctx, dir, "who.txt", marker)
				return dir
			}
			libA := newLib("A")
			libB := newLib("B")

			// mid carries a DIFFERENT repo under the same submodule name "lib".
			midRemote := filepath.Join(baseDir, "mid-remote")
			gitInitRepo(ctx, midRemote)
			gitAddSubmoduleSucceed(ctx, midRemote, libB, "lib")
			gitSucceed(ctx, midRemote, "commit", "-m", "add lib B")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, libA, "lib")
			gitAddSubmoduleSucceed(ctx, superRepo, midRemote, "mid")
			gitUpdateSubmodulesSucceed(ctx, superRepo, "--init", "--recursive")
			gitSucceed(ctx, superRepo, "commit", "-m", "add lib A and mid")

			expectReuseBlocked(ctx, superRepo, `is used more than once`)

			// Remotes stay reachable here, so the plain remote update must still produce correct
			// content for BOTH same-named submodules.
			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

			expectFileContent(filepath.Join(workTreeDir, "lib", "who.txt"), "A")
			expectFileContent(filepath.Join(workTreeDir, "mid", "lib", "who.txt"), "B")
		})

		// The everyday CI shape of an unserviceable store: it exists and is perfectly healthy, it
		// just predates the gitlink. Only looking the pinned object up can tell.
		It("blocks when the store exists but predates the pinned commit", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "first")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			// Move the submodule on in its remote and bump the gitlink to the new commit WITHOUT ever
			// fetching it into the superproject store.
			Expect(os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("second"), 0o644)).To(Succeed())
			gitSucceed(ctx, subRemote, "commit", "-am", "second")
			unfetchedSHA := headSHA(ctx, subRemote)
			gitSucceed(ctx, superRepo, "update-index", "--cacheinfo", "160000,"+unfetchedSHA+",sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "bump the gitlink past the store")

			storeDir := filepath.Join(superGitDir, "modules", "sub")
			Expect(storeDir).To(BeADirectory(), "the store must exist, or a cheaper check would explain the blocker")

			expectReuseBlocked(ctx, superRepo, "not initialized locally")

			// The remote does have the commit, so the plain remote update still completes.
			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
		})

		// Reuse enumerates nested submodules out of the pinned TREE, so a store holding the pinned
		// commit while missing that tree cannot serve the checkout either. Peeling the pin to its
		// tree is what tells the two apart.
		It("blocks when the store holds the pinned commit but not its tree", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "first")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			Expect(os.WriteFile(filepath.Join(subRemote, "file.txt"), []byte("second"), 0o644)).To(Succeed())
			gitSucceed(ctx, subRemote, "commit", "-am", "second")
			unfetchedSHA := headSHA(ctx, subRemote)
			gitSucceed(ctx, superRepo, "update-index", "--cacheinfo", "160000,"+unfetchedSHA+",sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "bump the gitlink past the store")

			// Plant the pinned COMMIT object alone into the store: a commit object names its tree but
			// does not carry it, so writing it back verbatim leaves the tree missing.
			storeDir := filepath.Join(superGitDir, "modules", "sub")
			commitObject := gitSucceed(ctx, subRemote, "cat-file", "commit", unfetchedSHA)
			planted := gitSucceedWithStdin(ctx, storeDir, commitObject, "hash-object", "-t", "commit", "-w", "--stdin")
			Expect(planted).To(Equal(unfetchedSHA), "the planted object must be the pin itself")

			hasCommit, err := gitObjectExists(ctx, storeDir, unfetchedSHA+"^{commit}")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasCommit).To(BeTrue(), "a commit-only probe would call this store serviceable")

			expectReuseBlocked(ctx, superRepo, "not initialized locally")

			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "second")
		})

		// protocol.file.allow=always is only safe while every URL of the invocation is one werf
		// computed, so a submodule that cannot be served locally must suppress the overrides
		// entirely — otherwise the relaxed transport would also un-block a file:// URL taken from
		// committed .gitmodules (CVE-2022-39253 class).
		It("keeps the file transport blocked for a submodule missing from the store", func(ctx SpecContext) {
			okRemote := filepath.Join(baseDir, "ok-remote")
			gitInitRepoWithFile(ctx, okRemote, "ok.txt", "OK")

			evilRemote := filepath.Join(baseDir, "evil-remote")
			gitInitRepoWithFile(ctx, evilRemote, "evil.txt", "EVIL")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, okRemote, "ok")
			gitAddSubmoduleSucceed(ctx, superRepo, "file://"+evilRemote, "evil")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodules")

			// Only "ok" is available locally: deinit "evil" so its module store no longer holds the pin.
			gitSucceed(ctx, superRepo, "submodule", "deinit", "-f", "evil")
			Expect(os.RemoveAll(filepath.Join(superGitDir, "modules", "evil"))).To(Succeed())

			expectReuseBlocked(ctx, superRepo, "not initialized locally")

			addWorkTree(ctx, headSHA(ctx, superRepo))

			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			err := updateSubmodules(ctx, superGitDir, workTreeDir)
			Expect(err).To(HaveOccurred(), "git must still refuse the file:// submodule")
			Expect(err.Error()).To(ContainSubstring("transport 'file' not allowed"))
			Expect(filepath.Join(workTreeDir, "evil", "evil.txt")).ToNot(BeAnExistingFile())
		})

		// A partial clone can hold the pinned commit while its blobs still live on the promisor
		// remote, so cloning from it would fail or need the very network reuse exists to avoid.
		DescribeTable("blocks on a partial-clone marker in the store",
			func(ctx SpecContext, configKey, configValue string) {
				subRemote := filepath.Join(baseDir, "sub-remote")
				gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

				gitInitRepo(ctx, superRepo)
				gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
				gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

				storeDir := filepath.Join(superGitDir, "modules", "sub")
				gitSucceed(ctx, storeDir, "config", configKey, configValue)

				expectReuseBlocked(ctx, superRepo, "local clone is partial")

				// The store is intact apart from the marker, so the plain remote update still works.
				allowFileTransport()
				addWorkTree(ctx, headSHA(ctx, superRepo))
				Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
				Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
				expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
			},
			Entry("extensions.partialclone", "extensions.partialclone", "origin"),
			Entry("remote.<name>.promisor", "remote.origin.promisor", "true"),
			Entry("remote.<name>.partialclonefilter", "remote.origin.partialclonefilter", "blob:none"),
		)

		// url.<store>.insteadOf=<remote> is keyed on the remote URL, so one URL shared by two
		// submodules would need two different stores behind the same key.
		It("blocks when two submodules share one remote URL", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "a")
			gitSucceed(ctx, superRepo, "-c", "protocol.file.allow=always", "submodule", "add", "--name", "b", subRemote, "b")
			gitSucceed(ctx, superRepo, "commit", "-m", "two submodules, one remote URL")

			// The shared URL is only reachable once the service worktree has its own submodule git
			// dirs, which is what records a remote to redirect: warm it, then let sync put the real
			// remote back the way every werf run does.
			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

			expectReuseBlocked(ctx, workTreeDir, "remote URL is shared with another submodule")

			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "a", "file.txt"), "hello")
			expectFileContent(filepath.Join(workTreeDir, "b", "file.txt"), "hello")
		})

		// insteadOf matches by prefix, so a remote URL that prefixes a store path would rewrite that
		// store URL too, and the clone would be pointed somewhere else entirely.
		It("blocks when a remote URL prefixes another submodule's store path", func(ctx SpecContext) {
			// The store path is <superRepo>/.git/modules/<name>, so a remote at <superRepo> minus its
			// suffix is a literal string prefix of it.
			superRepo = filepath.Join(baseDir, "pfx-super")
			superGitDir = filepath.Join(superRepo, ".git")
			subRemote := filepath.Join(baseDir, "pfx")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

			storeDir := filepath.Join(superGitDir, "modules", "sub")
			Expect(storeDir).To(HavePrefix(subRemote), "the spec is only meaningful while the URL really prefixes the store path")
			expectReuseBlocked(ctx, workTreeDir, "prefixes the object store path of")

			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
		})

		// Without a recorded remote there is nothing for url.<store>.insteadOf to key on, so the
		// already-populated submodule would keep fetching from wherever git resolves it.
		It("blocks when a populated submodule records no remote URL", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())

			worktreeGitDir, _, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
			Expect(err).ToNot(HaveOccurred())
			worktreeModuleDir := filepath.Join(worktreeGitDir, "modules", "sub")
			Expect(worktreeModuleDir).To(BeADirectory())
			gitSucceed(ctx, worktreeModuleDir, "config", "--unset", "remote.origin.url")

			expectReuseBlocked(ctx, workTreeDir, "has no remote URL to redirect")

			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
		})

		// The store path lands in the KEY of -c url.<path>.insteadOf, and git rejects a key carrying
		// an `=`. Reuse must notice that before handing such an option to git.
		It("blocks when the store path cannot be expressed as a git config key", func(ctx SpecContext) {
			superRepo = filepath.Join(baseDir, "super=eq")
			superGitDir = filepath.Join(superRepo, ".git")

			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			expectReuseBlocked(ctx, superRepo, "cannot be expressed as a git config key")

			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
		})

		// A repo-controlled name carrying an `=` would end the -c submodule.<name>.url key early, so
		// the store path would land in an unrelated option's value instead.
		It("blocks on a submodule whose committed name carries an equals sign", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			// git offers no way to commit such a name, so write .gitmodules directly. The store the
			// legitimate `submodule add` produced stays in place, so nothing else can block reuse.
			gitmodules := "[submodule \"ev=il\"]\n\tpath = sub\n\turl = " + subRemote + "\n"
			Expect(os.WriteFile(filepath.Join(superRepo, ".gitmodules"), []byte(gitmodules), 0o644)).To(Succeed())
			gitSucceed(ctx, superRepo, "add", ".gitmodules")
			gitSucceed(ctx, superRepo, "commit", "-m", "rename the submodule to a hostile name")

			expectReuseBlocked(ctx, superRepo, `unsupported submodule name "ev=il"`)

			allowFileTransport()
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
		})
	})

	// .gitmodules is committed content git validates in no way. Both of these must be ignored the
	// way git itself ignores them, and neither may cost the whole repository its reuse.
	Describe("hostile .gitmodules content that must not disable reuse", func() {
		// Reuse must come out of these unharmed: exactly the one real submodule redirected, and nothing
		// blocked. The store here is the main work tree's own, which is also the remote it records, so
		// only the submodule.<name>.url override applies.
		expectOnlySubIsRedirected := func(ctx context.Context) {
			overrides, blockers, err := submoduleLocalURLOverrides(ctx, superRepo)
			Expect(err).ToNot(HaveOccurred())
			Expect(blockers).To(BeEmpty())
			storeDir := filepath.Join(superGitDir, "modules", "sub")
			Expect(overrides).To(Equal([]string{
				"-c", "submodule.sub.url=" + storeDir,
				"-c", "url." + storeDir + ".insteadOf=" + filepath.Join(baseDir, "sub-remote"),
			}))
		}

		It("ignores an [include] section rather than reading it as a submodule", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			// `include.path` also ends in `.path`, so an unanchored key regex reads it as a submodule
			// named "include". Pointing it at the real submodule's path gives that phantom a gitlink,
			// which is what makes the misreading observable at all.
			gitmodules := gitSucceed(ctx, superRepo, "show", "HEAD:.gitmodules") + "[include]\n\tpath = sub\n"
			Expect(os.WriteFile(filepath.Join(superRepo, ".gitmodules"), []byte(gitmodules), 0o644)).To(Succeed())
			gitSucceed(ctx, superRepo, "add", ".gitmodules")
			gitSucceed(ctx, superRepo, "commit", "-m", "add an include section")

			expectOnlySubIsRedirected(ctx)

			Expect(os.RemoveAll(subRemote)).To(Succeed())
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
		})

		It("ignores an absolute and a .. submodule path rather than failing the update", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			// git refuses to add these itself; a repository can still ship them. As a pathspec each
			// would fail the ls-tree of the whole update, while git checks out no submodule there.
			gitmodules := gitSucceed(ctx, superRepo, "show", "HEAD:.gitmodules") +
				"[submodule \"abs\"]\n\tpath = " + filepath.Join(baseDir, "absolute-victim") + "\n\turl = https://werf-test-nonexistent.invalid/a.git\n" +
				"[submodule \"up\"]\n\tpath = ../escape\n\turl = https://werf-test-nonexistent.invalid/b.git\n"
			Expect(os.WriteFile(filepath.Join(superRepo, ".gitmodules"), []byte(gitmodules), 0o644)).To(Succeed())
			gitSucceed(ctx, superRepo, "add", ".gitmodules")
			gitSucceed(ctx, superRepo, "commit", "-m", "add non-local submodule paths")

			expectOnlySubIsRedirected(ctx)

			// Dropping the remote also pins that the surviving submodule is served locally: the plain
			// update this used to fall back to could not reach it any more.
			Expect(os.RemoveAll(subRemote)).To(Succeed())
			addWorkTree(ctx, headSHA(ctx, superRepo))
			Expect(syncSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			Expect(updateSubmodules(ctx, superGitDir, workTreeDir)).To(Succeed())
			expectFileContent(filepath.Join(workTreeDir, "sub", "file.txt"), "hello")
		})
	})

	Describe("discardWorktreeSubmoduleGitDirs", func() {
		// The fallback discards the service worktree's submodule state, and in a main work tree that
		// same path is the repository-wide submodule store. It must refuse rather than destroy data.
		It("refuses a main work tree", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")

			sharedStore := filepath.Join(superGitDir, "modules", "sub")
			Expect(sharedStore).To(BeADirectory())

			err := discardWorktreeSubmoduleGitDirs(ctx, superRepo)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a linked work tree"))
			Expect(sharedStore).To(BeADirectory(), "the repository-wide submodule store must survive")
			Expect(filepath.Join(superRepo, "sub", "file.txt")).To(BeARegularFile())
		})

		// .gitmodules is committed content, and git does not validate its paths, so the discard must
		// not delete at a path HEAD does not record as a gitlink.
		It("ignores a .gitmodules path without a gitlink", func(ctx SpecContext) {
			outside := filepath.Join(baseDir, "outside")
			Expect(os.MkdirAll(filepath.Join(outside, "inner"), 0o755)).To(Succeed())
			victim := filepath.Join(outside, "inner", "data.txt")
			Expect(os.WriteFile(victim, []byte("precious"), 0o644)).To(Succeed())

			gitInitRepo(ctx, superRepo)
			escape, err := filepath.Rel(superRepo, outside)
			Expect(err).ToNot(HaveOccurred())
			gitmodules := "[submodule \"escape\"]\n\tpath = " + escape + "\n\turl = https://werf-test-nonexistent.invalid/x.git\n"
			Expect(os.WriteFile(filepath.Join(superRepo, ".gitmodules"), []byte(gitmodules), 0o644)).To(Succeed())
			gitSucceed(ctx, superRepo, "add", ".gitmodules")
			gitSucceed(ctx, superRepo, "commit", "-m", "gitmodules entry without a gitlink")

			// The second variant reaches outside through a committed symlink used as an intermediate
			// path component, so it looks work-tree local and only the gitlink check rejects it.
			// RemoveAll would follow it, unlike a symlink in the final position, which it would
			// merely unlink.
			Expect(os.Symlink(outside, filepath.Join(superRepo, "link"))).To(Succeed())
			gitmodules += "[submodule \"vialink\"]\n\tpath = link/inner\n\turl = https://werf-test-nonexistent.invalid/y.git\n"
			Expect(os.WriteFile(filepath.Join(superRepo, ".gitmodules"), []byte(gitmodules), 0o644)).To(Succeed())
			gitSucceed(ctx, superRepo, "add", ".gitmodules", "link")
			gitSucceed(ctx, superRepo, "commit", "-m", "symlinked gitmodules path")

			addWorkTree(ctx, headSHA(ctx, superRepo))

			Expect(discardWorktreeSubmoduleGitDirs(ctx, workTreeDir)).To(Succeed())
			Expect(victim).To(BeARegularFile(), "a path outside the work tree must never be removed")
		})

		// The discard resolves what it deletes from the work tree's own .git pointer, not from
		// `git rev-parse`, which an inherited GIT_DIR silently redirects at another repository.
		It("removes only the given work tree's submodule git dirs under an inherited GIT_DIR", func(ctx SpecContext) {
			subRemote := filepath.Join(baseDir, "sub-remote")
			gitInitRepoWithFile(ctx, subRemote, "file.txt", "hello")

			gitInitRepo(ctx, superRepo)
			gitAddSubmoduleSucceed(ctx, superRepo, subRemote, "sub")
			gitSucceed(ctx, superRepo, "commit", "-m", "add submodule")
			commit := headSHA(ctx, superRepo)

			bystanderWorkTree := filepath.Join(baseDir, "bystander")
			gitSucceed(ctx, superRepo, "worktree", "add", "--detach", bystanderWorkTree, commit)
			addWorkTree(ctx, commit)
			for _, dir := range []string{bystanderWorkTree, workTreeDir} {
				gitUpdateSubmodulesSucceed(ctx, dir, "--init")
			}

			targetGitDir, _, err := resolveWorkTreeGitDirs(ctx, workTreeDir)
			Expect(err).ToNot(HaveOccurred())
			bystanderGitDir, _, err := resolveWorkTreeGitDirs(ctx, bystanderWorkTree)
			Expect(err).ToNot(HaveOccurred())
			bystanderModules := filepath.Join(bystanderGitDir, "modules", "sub")
			Expect(bystanderModules).To(BeADirectory())
			Expect(filepath.Join(targetGitDir, "modules", "sub")).To(BeADirectory())

			// This is what defeated the earlier `git rev-parse`-based resolution: every git
			// invocation, and the discard with it, would have answered for the bystander instead.
			setEnvForSpec("GIT_DIR", bystanderGitDir)

			Expect(discardWorktreeSubmoduleGitDirs(ctx, workTreeDir)).To(Succeed())

			Expect(bystanderModules).To(BeADirectory(), "another work tree's submodule git dirs must survive")
			Expect(filepath.Join(bystanderWorkTree, "sub", "file.txt")).To(BeARegularFile())
			Expect(filepath.Join(targetGitDir, "modules")).ToNot(BeAnExistingFile())
			Expect(filepath.Join(workTreeDir, "sub")).ToNot(BeAnExistingFile())
		})
	})
})
