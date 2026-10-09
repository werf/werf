package git_repo_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/git_repo/gitdata"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("LFS repository read and patch preparation", func() {
	ginkgo.DescribeTable("does not download excluded LFS objects", func(ctx ginkgo.SpecContext, operation string) {
		root, err := filepath.EvalSymlinks(ginkgo.GinkgoT().TempDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "global.gitconfig"))
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "1")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_KEY_0", "protocol.file.allow")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_VALUE_0", "always")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_PARAMETERS", "")
		ginkgo.GinkgoT().Setenv("GIT_LFS_SKIP_SMUDGE", "0")
		ginkgo.GinkgoT().Setenv("GIT_LFS_SKIP_DOWNLOAD_ERRORS", "0")
		gomega.Expect(werf.Init(root, filepath.Join(root, "werf-home"))).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		manager, err := gitdata.GetHostGitDataManager(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(git_repo.Init(manager)).To(gomega.Succeed())
		runGit := func(dir string, args ...string) {
			utils.RunSucceedCommand(ctx, dir, "git", append([]string{"-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		}
		oldDir, newDir, sourceDir := filepath.Join(root, "old"), filepath.Join(root, "new"), filepath.Join(root, "source")
		for _, fixture := range []struct{ dir, contents string }{{oldDir, "old\n"}, {newDir, "new\n"}} {
			gomega.Expect(os.Mkdir(fixture.dir, 0o755)).To(gomega.Succeed())
			runGit(fixture.dir, "init", "--initial-branch=main")
			gomega.Expect(os.WriteFile(filepath.Join(fixture.dir, "plain.txt"), []byte(fixture.contents), 0o644)).To(gomega.Succeed())
			runGit(fixture.dir, "add", ".")
			runGit(fixture.dir, "commit", "-m", "submodule content")
		}
		gomega.Expect(os.Mkdir(sourceDir, 0o755)).To(gomega.Succeed())
		runGit(sourceDir, "init", "--initial-branch=main")
		gomega.Expect(os.WriteFile(filepath.Join(sourceDir, ".gitattributes"), []byte("*.bin filter=lfs diff=lfs merge=lfs -text\n"), 0o644)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(filepath.Join(sourceDir, "excluded.bin"), []byte("version https://git-lfs.github.com/spec/v1\noid sha256:"+strings.Repeat("0", 64)+"\nsize 7\n"), 0o644)).To(gomega.Succeed())
		var commits []string
		for _, fixture := range []struct{ dir, contents string }{{oldDir, "before\n"}, {newDir, "after\n"}} {
			gomega.Expect(os.WriteFile(filepath.Join(sourceDir, ".gitmodules"), []byte("[submodule \"sub\"]\npath = sub\nurl = "+filepath.ToSlash(fixture.dir)+"\n"), 0o644)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(sourceDir, "wanted.txt"), []byte(fixture.contents), 0o644)).To(gomega.Succeed())
			runGit(sourceDir, "add", ".gitattributes", ".gitmodules", "excluded.bin", "wanted.txt")
			runGit(sourceDir, "update-index", "--add", "--cacheinfo", "160000,"+utils.GetHeadCommit(ctx, fixture.dir)+",sub")
			runGit(sourceDir, "commit", "-m", "superproject content")
			commits = append(commits, utils.GetHeadCommit(ctx, sourceDir))
		}
		runGit(sourceDir, "remote", "add", "origin", "file://"+filepath.ToSlash(sourceDir))
		runGit(sourceDir, "lfs", "install", "--skip-repo")
		if operation == "remote read" {
			repo, err := git_repo.OpenRemoteRepo("lfs-read", sourceDir, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repo.CloneAndFetch(ctx)).To(gomega.Succeed())
			contents, err := repo.ReadCommitFile(ctx, commits[1], "wanted.txt")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(contents).To(gomega.Equal([]byte("after\n")))
			return
		}
		repo, err := git_repo.OpenLocalRepo(ctx, "lfs-patch-recovery", sourceDir, git_repo.OpenLocalRepoOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		if operation == "local read" {
			contents, err := repo.ReadCommitFile(ctx, commits[1], "wanted.txt")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(contents).To(gomega.Equal([]byte("after\n")))
			return
		}
		cacheDir := filepath.Join(root, "worktree-cache")
		gitDir := filepath.Join(sourceDir, ".git")
		matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"wanted.txt", "sub/plain.txt"}})
		opts := git_repo.PatchOptions{FromCommit: commits[0], ToCommit: commits[1], PathMatcher: matcher}
		var initial bytes.Buffer
		_, err = true_git.Patch(ctx, &initial, gitDir, filepath.Join(root, "probe-cache"), true, true_git.PatchOptions(opts))
		gomega.Expect(true_git.IsCommitsNotPresentError(err)).To(gomega.BeTrue(), "fixture must exercise the missing-submodule-commit recovery: %v", err)
		patch, err := repo.CreatePatch(ctx, sourceDir, gitDir, "lfs-patch-recovery", cacheDir, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(patch.GetPaths()).To(gomega.ConsistOf("wanted.txt", "sub/plain.txt"))
		contents, err := os.ReadFile(patch.GetFilePath())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		for _, line := range []string{"-before", "+after", "-old", "+new"} {
			gomega.Expect(string(contents)).To(gomega.ContainSubstring(line))
		}
	},
		ginkgo.Entry("local committed-file read", "local read"),
		ginkgo.Entry("remote committed-file read", "remote read"),
		ginkgo.Entry("missing source submodule commit recovery", "patch recovery"),
	)
})
