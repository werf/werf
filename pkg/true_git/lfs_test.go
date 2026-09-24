package true_git

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("Git LFS", func() {
	const pointer = "version https://git-lfs.github.com/spec/v1\noid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393\nsize 12345\n"

	DescribeTable("detecting pointer files",
		func(content string, expected bool) {
			Expect(isLfsPointer([]byte(content))).To(Equal(expected))
		},
		Entry("spec v1 pointer", pointer, true),
		Entry("pre-release hawser pointer", "version https://hawser.github.com/spec/v1\noid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393\nsize 12345\n", true),
		Entry("pointer with extension line", "version https://git-lfs.github.com/spec/v1\next-0-foo sha256:abc\noid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393\nsize 12345\n", true),
		Entry("empty file", "", false),
		Entry("plain text", "hello\nworld\n", false),
		Entry("version line without oid", "version https://git-lfs.github.com/spec/v1\nsize 12345\n", false),
		Entry("prose mentioning the spec url", "see version https://git-lfs.github.com/spec/v1 for details\noid sha256:abc\n", false),
		Entry("oversized content", pointer+strings.Repeat("x", lfsPointerMaxSize), false),
	)

	It("changes the archive ID only when lfs is enabled", func() {
		opts := ArchiveOptions{
			Commit:      "0123456789012345678901234567890123456789",
			PathScope:   "assets",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "assets"}),
		}
		plain := opts.ID()

		opts.LfsEnv = []string{"GIT_ASKPASS=/tmp/askpass"}
		Expect(opts.ID()).To(Equal(plain))

		opts.Lfs = true
		Expect(opts.ID()).NotTo(Equal(plain))
	})

	Describe("archiving a repository with LFS-tracked files", func() {
		const objectContent = "REAL-BINARY-CONTENT"

		var originDir, cloneDir, workTreeCacheDir string

		lfsFilterOpts := []string{
			"-c", "filter.lfs.smudge=git-lfs smudge -- %f",
			"-c", "filter.lfs.process=git-lfs filter-process",
			"-c", "filter.lfs.clean=git-lfs clean -- %f",
			"-c", "filter.lfs.required=true",
		}

		archive := func(ctx SpecContext, lfs bool) map[string]string {
			var buf bytes.Buffer
			err := Archive(ctx, &buf, filepath.Join(cloneDir, ".git"), workTreeCacheDir, ArchiveOptions{
				Commit:      utils.GetHeadCommit(ctx, cloneDir),
				PathScope:   "assets",
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "assets"}),
				Lfs:         lfs,
			})
			Expect(err).To(Succeed())
			return readTarFiles(&buf)
		}

		BeforeEach(func(ctx SpecContext) {
			if _, err := exec.LookPath("git-lfs"); err != nil {
				Skip("git-lfs is not installed")
			}
			// No global config: the host must not have `git lfs install` filters for the spec to
			// prove werf supplies them itself.
			isolateGitConfig()

			originDir = filepath.Join(SuiteData.TestDirPath, "origin")
			cloneDir = filepath.Join(SuiteData.TestDirPath, "clone")
			workTreeCacheDir = filepath.Join(SuiteData.TestDirPath, "worktree")

			gitInitRepo(ctx, originDir)
			utils.MkdirAll(filepath.Join(originDir, "assets"))
			Expect(os.WriteFile(filepath.Join(originDir, ".gitattributes"), []byte("*.bin filter=lfs diff=lfs merge=lfs -text\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(originDir, "assets", "big.bin"), []byte(objectContent), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(originDir, "assets", "plain.txt"), []byte("plain\n"), 0o644)).To(Succeed())
			gitSucceed(ctx, originDir, append(lfsFilterOpts, "add", ".")...)
			gitSucceed(ctx, originDir, append(lfsFilterOpts, "commit", "-m", "lfs content")...)

			// A plain clone over file:// stores pointer files only and uses the origin as LFS endpoint.
			gitSucceed(ctx, SuiteData.TestDirPath, "clone", "-q", "file://"+originDir, cloneDir)
			Expect(isLfsPointer(lo.Must(os.ReadFile(filepath.Join(cloneDir, "assets", "big.bin"))))).To(BeTrue())

			Expect(werf.Init(GinkgoT().TempDir(), GinkgoT().TempDir())).Should(Succeed())
			Expect(Init(ctx, Options{})).Should(Succeed())
		})

		It("archives pointer files when lfs is disabled", func(ctx SpecContext) {
			files := archive(ctx, false)
			Expect(files).To(HaveKey("plain.txt"))
			Expect(isLfsPointer([]byte(files["big.bin"]))).To(BeTrue())
		})

		It("archives the object content when lfs is enabled", func(ctx SpecContext) {
			files := archive(ctx, true)
			Expect(files["plain.txt"]).To(Equal("plain\n"))
			Expect(files["big.bin"]).To(Equal(objectContent))
		})

		It("fails when an object cannot be fetched", func(ctx SpecContext) {
			Expect(os.RemoveAll(filepath.Join(originDir, ".git", "lfs"))).To(Succeed())

			var buf bytes.Buffer
			err := Archive(ctx, &buf, filepath.Join(cloneDir, ".git"), workTreeCacheDir, ArchiveOptions{
				Commit:      utils.GetHeadCommit(ctx, cloneDir),
				PathScope:   "assets",
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "assets"}),
				Lfs:         true,
			})
			Expect(err).To(MatchError(ContainSubstring("git lfs pull command failed")))
		})

		It("fails when a pointer file is left behind by a successful pull", func(ctx SpecContext) {
			// Without the LFS attribute git-lfs no longer treats big.bin as its file, so the pull
			// succeeds and leaves the pointer blob in place.
			gitSucceed(ctx, cloneDir, "rm", "-q", ".gitattributes")
			gitSucceed(ctx, cloneDir, "commit", "-m", "untrack")

			var buf bytes.Buffer
			err := Archive(ctx, &buf, filepath.Join(cloneDir, ".git"), workTreeCacheDir, ArchiveOptions{
				Commit:      utils.GetHeadCommit(ctx, cloneDir),
				PathScope:   "assets",
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "assets"}),
				Lfs:         true,
			})
			Expect(err).To(MatchError(ContainSubstring(`file "assets/big.bin" is still a Git LFS pointer`)))
		})
	})
})

func readTarFiles(r io.Reader) map[string]string {
	files := map[string]string{}
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return files
		}
		Expect(err).To(Succeed())
		if header.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		Expect(err).To(Succeed())
		files[header.Name] = string(data)
	}
}
