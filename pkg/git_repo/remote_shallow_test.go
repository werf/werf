package git_repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("Remote shallow mirror", func() {
	var sourceDir string
	var sourceURL string

	openRemote := func(branch, tag, commit string) *Remote {
		repo, err := OpenRemoteRepo("test-mapping", sourceURL, nil)
		Expect(err).NotTo(HaveOccurred())
		repo.Branch = branch
		repo.Tag = tag
		repo.Commit = commit
		return repo
	}

	gitInSource := func(ctx SpecContext, args ...string) {
		utils.RunSucceedCommand(ctx, sourceDir, "git", append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	}

	BeforeEach(func(ctx SpecContext) {
		tmpDir := GinkgoT().TempDir()
		Expect(werf.Init(tmpDir, GinkgoT().TempDir())).To(Succeed())
		Expect(true_git.Init(ctx, true_git.Options{})).To(Succeed())
		Expect(Init(&fakeGitDataManager{})).To(Succeed())

		resetLsRemoteTagsCache()

		sourceDir = filepath.Join(tmpDir, "source")
		Expect(os.MkdirAll(sourceDir, 0o755)).To(Succeed())
		sourceURL = sourceDir

		utils.RunSucceedCommand(ctx, sourceDir, "git", "-c", "init.defaultBranch=main", "init")
		gitInSource(ctx, "checkout", "-b", "main")
		gitInSource(ctx, "commit", "--allow-empty", "-m", "c1")
		gitInSource(ctx, "commit", "--allow-empty", "-m", "c2")
	})

	Describe("mirror kind selection", func() {
		DescribeTable("resolveMirrorKind",
			func(branch, tag, commit string, expected mirrorKind) {
				repo := openRemote(branch, tag, commit)
				kind, err := repo.resolveMirrorKind()
				Expect(err).NotTo(HaveOccurred())
				Expect(kind).To(Equal(expected))
			},
			Entry("branch mapping uses full", "main", "", "", mirrorKindFull),
			Entry("no refs uses full", "", "", "", mirrorKindFull),
			Entry("tag mapping uses shallow", "", "v1", "", mirrorKindShallow),
			Entry("commit mapping uses shallow", "", "", "0123456789012345678901234567890123456789", mirrorKindShallow),
		)

		It("uses full when requires_full marker is set", func() {
			repo := openRemote("", "v1", "")
			Expect(os.MkdirAll(filepath.Dir(repo.requiresFullMarkerPath()), 0o755)).To(Succeed())
			Expect(repo.writeRequiresFullMarker()).To(Succeed())

			kind, err := repo.resolveMirrorKind()
			Expect(err).NotTo(HaveOccurred())
			Expect(kind).To(Equal(mirrorKindFull))
		})
	})

	Describe("full mirror flows", func() {
		It("clones and fetches branch mapping into full mirror", func(ctx SpecContext) {
			repo := openRemote("main", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			Expect(repo.GetClonePath()).To(HaveSuffix(string(mirrorKindFull)))
			Expect(repo.GetClonePath()).To(BeADirectory())

			headCommit := utils.GetHeadCommit(ctx, sourceDir)
			branchCommit, err := repo.LatestBranchCommit(ctx, "main")
			Expect(err).NotTo(HaveOccurred())
			Expect(branchCommit).To(Equal(headCommit))
		})

		It("clones refless remote (includes-style) into full mirror", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")

			repo := openRemote("", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			Expect(repo.GetClonePath()).To(HaveSuffix(string(mirrorKindFull)))

			headCommit := utils.GetHeadCommit(ctx, sourceDir)
			tagCommit, err := repo.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(tagCommit).To(Equal(headCommit))
		})
	})

	Describe("TagCommit peel", func() {
		setupTags := func(ctx SpecContext) {
			gitInSource(ctx, "tag", "light")
			gitInSource(ctx, "tag", "-a", "annot", "-m", "annotated")
			gitInSource(ctx, "-c", "advice.nestedTag=false", "tag", "-a", "nested", "-m", "nested", "annot")
		}

		DescribeTable("resolves tag chains to commit in full mirror",
			func(ctx SpecContext, tag string) {
				setupTags(ctx)

				repo := openRemote("", "", "")
				Expect(repo.CloneAndFetch(ctx)).To(Succeed())

				headCommit := utils.GetHeadCommit(ctx, sourceDir)
				tagCommit, err := repo.TagCommit(ctx, tag)
				Expect(err).NotTo(HaveOccurred())
				Expect(tagCommit).To(Equal(headCommit))
			},
			Entry("lightweight tag", "light"),
			Entry("annotated tag", "annot"),
			Entry("nested annotated tag chain", "nested"),
		)

		DescribeTable("resolves tag chains to commit in shallow mirror",
			func(ctx SpecContext, tag string) {
				setupTags(ctx)

				repo := openRemote("", tag, "")
				Expect(repo.CloneAndFetch(ctx)).To(Succeed())
				Expect(repo.GetClonePath()).To(HaveSuffix(string(mirrorKindShallow)))

				headCommit := utils.GetHeadCommit(ctx, sourceDir)
				tagCommit, err := repo.TagCommit(ctx, tag)
				Expect(err).NotTo(HaveOccurred())
				Expect(tagCommit).To(Equal(headCommit))
			},
			Entry("lightweight tag", "light"),
			Entry("annotated tag", "annot"),
			Entry("nested annotated tag chain", "nested"),
		)

		It("returns bad tag error for missing tag in full mirror", func(ctx SpecContext) {
			repo := openRemote("", "", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			_, err := repo.TagCommit(ctx, "nonexistent")
			Expect(err).To(MatchError(ContainSubstring("bad tag")))
		})
	})

	Describe("shallow commit mapping", func() {
		It("rejects non-full-length commit SHA", func(ctx SpecContext) {
			repo := openRemote("", "", "abc")
			repo.kind = mirrorKindShallow
			err := repo.cloneAndFetchShallow(ctx)
			Expect(err).To(MatchError(ContainSubstring("full-length commit SHA required")))
		})

		It("fetches commit shallowly and skips network when already present", func(ctx SpecContext) {
			headCommit := utils.GetHeadCommit(ctx, sourceDir)

			repo := openRemote("", "", headCommit)
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())
			Expect(repo.GetClonePath()).To(HaveSuffix(string(mirrorKindShallow)))

			exists, err := repo.IsCommitExists(ctx, headCommit)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			Expect(os.RemoveAll(sourceDir)).To(Succeed())

			repo2 := openRemote("", "", headCommit)
			Expect(repo2.CloneAndFetch(ctx)).To(Succeed())

			exists, err = repo2.IsCommitExists(ctx, headCommit)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})

	Describe("shallow tag mapping", func() {
		It("fetches tag shallowly with depth=1", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")
			headCommit := utils.GetHeadCommit(ctx, sourceDir)

			repo := openRemote("", "v1", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			tagCommit, err := repo.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(tagCommit).To(Equal(headCommit))

			isShallow, err := true_git.IsShallowClone(ctx, repo.GetClonePath())
			Expect(err).NotTo(HaveOccurred())
			Expect(isShallow).To(BeTrue())
		})

		It("skips fetch when tag already resolves locally", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")
			headCommit := utils.GetHeadCommit(ctx, sourceDir)

			repo := openRemote("", "v1", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			Expect(os.RemoveAll(sourceDir)).To(Succeed())

			repo2 := openRemote("", "v1", "")
			Expect(repo2.CloneAndFetch(ctx)).To(Succeed())

			tagCommit, err := repo2.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(tagCommit).To(Equal(headCommit))
		})

		It("updates stale tag ref without fetch when commit is already present", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v2")
			gitInSource(ctx, "tag", "v1", "HEAD~1")
			headCommit := utils.GetHeadCommit(ctx, sourceDir)

			repoV2 := openRemote("", "v2", "")
			Expect(repoV2.CloneAndFetch(ctx)).To(Succeed())

			repoV1 := openRemote("", "v1", "")
			Expect(repoV1.CloneAndFetch(ctx)).To(Succeed())
			prevCommit, err := repoV1.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(prevCommit).NotTo(Equal(headCommit))

			gitInSource(ctx, "tag", "-f", "v1", headCommit)
			resetLsRemoteTagsCache()

			repoV1Again := openRemote("", "v1", "")
			Expect(repoV1Again.CloneAndFetch(ctx)).To(Succeed())
			movedCommit, err := repoV1Again.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(movedCommit).To(Equal(headCommit))
		})

		It("fetches again after retag to a new commit", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")

			repo := openRemote("", "v1", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			gitInSource(ctx, "commit", "--allow-empty", "-m", "c3")
			gitInSource(ctx, "tag", "-f", "v1")
			newHeadCommit := utils.GetHeadCommit(ctx, sourceDir)

			resetLsRemoteTagsCache()

			repo2 := openRemote("", "v1", "")
			Expect(repo2.CloneAndFetch(ctx)).To(Succeed())

			tagCommit, err := repo2.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(tagCommit).To(Equal(newHeadCommit))
		})

		It("returns bad tag error when tag is missing in remote", func(ctx SpecContext) {
			repo := openRemote("", "nonexistent", "")
			err := repo.CloneAndFetch(ctx)
			Expect(err).To(MatchError(ContainSubstring("bad tag")))
			Expect(err).To(MatchError(ContainSubstring("not found in remote")))

			_, statErr := os.Stat(repo.requiresFullMarkerPath())
			Expect(os.IsNotExist(statErr)).To(BeTrue())
		})

		It("caches ls-remote results per URL within the process", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")
			gitInSource(ctx, "tag", "v2")

			repo := openRemote("", "v1", "")
			Expect(repo.CloneAndFetch(ctx)).To(Succeed())

			lsRemoteTagsCacheMu.Lock()
			lenCache := len(lsRemoteTagsCache)
			lsRemoteTagsCacheMu.Unlock()
			Expect(lenCache).To(Equal(1))
		})
	})

	Describe("fallback to full mirror", func() {
		It("keeps sibling shallow mappings usable after submodule downgrade", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")
			v1Commit := utils.GetHeadCommit(ctx, sourceDir)

			shallowRepo := openRemote("", "v1", "")
			Expect(shallowRepo.CloneAndFetch(ctx)).To(Succeed())
			_, err := shallowRepo.initRepoHandleBackedByWorkTree(ctx, v1Commit)
			Expect(err).NotTo(HaveOccurred())
			shallowPath := shallowRepo.GetClonePath()
			shallowWorktreePath := shallowRepo.getWorkTreeCacheDir(shallowRepo.getRepoID())

			gitmodulesPath := filepath.Join(sourceDir, ".gitmodules")
			Expect(os.WriteFile(gitmodulesPath, []byte("[submodule \"sub\"]\n\tpath = sub\n\turl = ../sub\n"), 0o644)).To(Succeed())
			gitInSource(ctx, "add", ".gitmodules")
			gitInSource(ctx, "commit", "-m", "add submodule")
			gitInSource(ctx, "tag", "v2")
			v2Commit := utils.GetHeadCommit(ctx, sourceDir)
			resetLsRemoteTagsCache()

			fullRepo := openRemote("", "v2", "")
			Expect(fullRepo.CloneAndFetch(ctx)).To(Succeed())
			Expect(fullRepo.GetClonePath()).To(HaveSuffix(string(mirrorKindFull)))
			Expect(fullRepo.GetClonePath()).To(BeADirectory())
			Expect(fullRepo.requiresFullMarkerPath()).To(BeARegularFile())
			Expect(shallowPath).To(BeADirectory())
			Expect(shallowWorktreePath).To(BeADirectory())
			_, err = shallowRepo.initRepoHandleBackedByWorkTree(ctx, v1Commit)
			Expect(err).NotTo(HaveOccurred())

			v1TagCommit, err := shallowRepo.TagCommit(ctx, "v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(v1TagCommit).To(Equal(v1Commit))

			v2TagCommit, err := fullRepo.TagCommit(ctx, "v2")
			Expect(err).NotTo(HaveOccurred())
			Expect(v2TagCommit).To(Equal(v2Commit))

			nextRepo := openRemote("", "v1", "")
			Expect(nextRepo.CloneAndFetch(ctx)).To(Succeed())
			Expect(nextRepo.GetClonePath()).To(HaveSuffix(string(mirrorKindFull)))
		})

		It("does not persist marker when remote is unreachable", func(ctx SpecContext) {
			repo := openRemote("", "v1", "")
			Expect(os.RemoveAll(sourceDir)).To(Succeed())

			err := repo.CloneAndFetch(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("underlying shallow error"))

			_, statErr := os.Stat(repo.requiresFullMarkerPath())
			Expect(os.IsNotExist(statErr)).To(BeTrue())
		})

		It("does not persist marker for a well-formed but nonexistent commit", func(ctx SpecContext) {
			repo := openRemote("", "", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

			err := repo.CloneAndFetch(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found in full mirror after fallback"))

			_, statErr := os.Stat(repo.requiresFullMarkerPath())
			Expect(os.IsNotExist(statErr)).To(BeTrue())
		})
	})

	Describe("buildFetchOptions", func() {
		It("fetches only the mapped branch and no tags for a branch mapping", func() {
			opts := buildFetchOptions("origin", "main")
			Expect(opts.Tags).To(Equal(git.NoTags))
			Expect(opts.RefSpecs).To(BeEmpty())
		})

		It("fetches all branches and tags when no branch is set so off-branch commits are reachable after fallback", func() {
			opts := buildFetchOptions("origin", "")
			Expect(opts.Tags).To(Equal(git.AllTags))
			Expect(opts.RefSpecs).To(ConsistOf(gitconfig.RefSpec("+refs/heads/*:refs/remotes/origin/*")))
		})
	})

	Describe("isBySHAFetchRefusal", func() {
		DescribeTable("recognizes capability refusal patterns",
			func(msg string, expected bool) {
				Expect(isBySHAFetchRefusal(errors.New(msg))).To(Equal(expected))
			},
			Entry("unadvertised object", "upload-pack: not our ref, Server does not allow request for unadvertised object", true),
			Entry("not our ref", "fatal: remote error: upload-pack: not our ref deadbeef", true),
			Entry("shallow unsupported", "server does not support shallow requests", true),
			Entry("network error", "fatal: unable to access 'https://example.com/': Could not resolve host", false),
			Entry("auth error", "fatal: Authentication failed", false),
		)
	})

	Describe("concurrent access", func() {
		It("serializes concurrent clone and fetch of the same URL", func(ctx SpecContext) {
			gitInSource(ctx, "tag", "v1")

			var wg sync.WaitGroup
			errs := make([]error, 4)
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					defer GinkgoRecover()
					var repo *Remote
					var err error
					if i%2 == 0 {
						repo, err = OpenRemoteRepo(fmt.Sprintf("branch-mapping-%d", i), sourceURL, nil)
						if err == nil {
							repo.Branch = "main"
						}
					} else {
						repo, err = OpenRemoteRepo(fmt.Sprintf("tag-mapping-%d", i), sourceURL, nil)
						if err == nil {
							repo.Tag = "v1"
						}
					}
					if err != nil {
						errs[i] = err
						return
					}
					errs[i] = repo.CloneAndFetch(ctx)
				}(i)
			}
			wg.Wait()

			for i, err := range errs {
				Expect(err).NotTo(HaveOccurred(), "goroutine %d", i)
			}
		})
	})
})
