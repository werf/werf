package git_repo_test

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

// Shallow submodule clones must be invisible in everything werf derives from a commit: the
// checksum, the archive and the patch between two commits whose gitlink moved — including the
// patch retry werf makes when the submodule lacks the source commit. Each variant works on its own
// clone of the superproject, so nothing one of them fetched or cached can serve the other.
var _ = ginkgo.Describe("git data of a superproject with shallow submodule clones", func() {
	ginkgo.It("matches the git data produced with full submodule clones", func(ctx ginkgo.SpecContext) {
		root, err := filepath.EvalSymlinks(ginkgo.GinkgoT().TempDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "global.gitconfig"))
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "1")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_KEY_0", "protocol.file.allow")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_VALUE_0", "always")
		ginkgo.GinkgoT().Setenv("GIT_CONFIG_PARAMETERS", "")
		gomega.Expect(werf.Init(root, filepath.Join(root, "werf-home"))).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		manager, err := gitdata.GetHostGitDataManager(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(git_repo.Init(manager)).To(gomega.Succeed())
		runGit := func(dir string, args ...string) {
			utils.RunSucceedCommand(ctx, dir, "git", append([]string{"-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		}
		commitFile := func(dir, name, contents string) string {
			gomega.Expect(os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644)).To(gomega.Succeed())
			runGit(dir, "add", name)
			runGit(dir, "commit", "-m", contents)
			return utils.GetHeadCommit(ctx, dir)
		}

		// The submodule has history before, between and after the two commits the superproject
		// pins, so a full clone and a shallow one differ in what they hold.
		subDir := filepath.Join(root, "sub")
		gomega.Expect(os.Mkdir(subDir, 0o755)).To(gomega.Succeed())
		runGit(subDir, "init", "--initial-branch=main")
		commitFile(subDir, "plain.txt", "ancient\n")
		subOld := commitFile(subDir, "plain.txt", "old\n")
		commitFile(subDir, "plain.txt", "between\n")
		subNew := commitFile(subDir, "plain.txt", "new\n")
		commitFile(subDir, "plain.txt", "tip\n")

		sourceDir := filepath.Join(root, "source")
		gomega.Expect(os.Mkdir(sourceDir, 0o755)).To(gomega.Succeed())
		runGit(sourceDir, "init", "--initial-branch=main")
		gomega.Expect(os.WriteFile(filepath.Join(sourceDir, ".gitmodules"), []byte("[submodule \"sub\"]\n\tpath = sub\n\turl = file://"+filepath.ToSlash(subDir)+"\n"), 0o644)).To(gomega.Succeed())
		runGit(sourceDir, "add", ".gitmodules")
		var commits []string
		for _, pin := range []string{subOld, subNew} {
			runGit(sourceDir, "update-index", "--add", "--cacheinfo", "160000,"+pin+",sub")
			runGit(sourceDir, "commit", "-m", "pin "+pin)
			commits = append(commits, utils.GetHeadCommit(ctx, sourceDir))
		}

		matcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"sub/plain.txt"}})
		type gitData struct {
			checksum string
			archive  map[string]string
			patch    string
			paths    []string
		}
		collect := func(variant string) gitData {
			// A clone holds no submodule objects of its own, so the service worktree has to fetch
			// them from the submodule's remote: the path shallow clones apply to.
			cloneDir := filepath.Join(root, variant)
			runGit(root, "clone", "-q", sourceDir, cloneDir)
			repo, err := git_repo.OpenLocalRepo(ctx, variant, cloneDir, git_repo.OpenLocalRepoOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			checksum, err := repo.GetOrCreateChecksum(ctx, git_repo.ChecksumOptions{LsTreeOptions: git_repo.LsTreeOptions{PathMatcher: matcher}, Commit: commits[1]})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			archive, err := repo.GetOrCreateArchive(ctx, git_repo.ArchiveOptions{Commit: commits[1], PathMatcher: matcher})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			archiveData, err := os.ReadFile(archive.GetFilePath())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			patch, err := repo.GetOrCreatePatch(ctx, git_repo.PatchOptions{FromCommit: commits[0], ToCommit: commits[1], PathMatcher: matcher})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			patchData, err := os.ReadFile(patch.GetFilePath())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			return gitData{checksum: checksum, archive: readTar(archiveData), patch: string(patchData), paths: patch.GetPaths()}
		}

		shallow := collect("shallow")
		ginkgo.GinkgoT().Setenv("WERF_DISABLE_SHALLOW_SUBMODULES", "1")
		full := collect("full")

		gomega.Expect(submoduleCloneMarkers(filepath.Join(root, "shallow"))).To(gomega.ConsistOf(true), "the shallow variant must have produced exactly one shallow submodule clone")
		gomega.Expect(submoduleCloneMarkers(filepath.Join(root, "full"))).To(gomega.ConsistOf(false), "the full variant must have produced exactly one full submodule clone")

		gomega.Expect(shallow.archive).To(gomega.HaveKeyWithValue("sub/plain.txt", "new\n"))
		gomega.Expect(shallow.paths).To(gomega.ConsistOf("sub/plain.txt"))
		gomega.Expect(shallow.patch).To(gomega.ContainSubstring("-old\n"))
		gomega.Expect(shallow.patch).To(gomega.ContainSubstring("+new\n"))
		gomega.Expect(shallow).To(gomega.Equal(full))
	})
})

func readTar(data []byte) map[string]string {
	files := make(map[string]string)
	reader := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return files
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		if header.Typeflag != tar.TypeReg {
			continue
		}
		content, err := io.ReadAll(reader)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		files[header.Name] = string(content)
	}
}

// submoduleCloneMarkers reports, for every submodule git dir of the service worktrees werf linked
// to the repository at repoDir, whether that git dir is shallow.
func submoduleCloneMarkers(repoDir string) []bool {
	var markers []bool
	err := filepath.WalkDir(filepath.Join(repoDir, ".git", "worktrees"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() || filepath.Base(path) != "sub" || filepath.Base(filepath.Dir(path)) != "modules" {
			return nil
		}
		_, err = os.Stat(filepath.Join(path, "shallow"))
		markers = append(markers, err == nil)
		return filepath.SkipDir
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return markers
}
