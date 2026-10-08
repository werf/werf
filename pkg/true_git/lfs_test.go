package true_git

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/test/pkg/suite_init"
)

var _ = ginkgo.Describe("automatic Git LFS archives", func() {
	suite_init.NewWerfInitData(SuiteData.TmpDirData)
	var repoDir, cacheDir, commit string
	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		isolateGitConfig()
		setEnvForSpec("GIT_TERMINAL_PROMPT", "0")
		setEnvForSpec("GIT_ASKPASS", "false")
		setEnvForSpec("SSH_ASKPASS", "false")
		_, err := exec.LookPath("git-lfs")
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Git LFS archive specs require git-lfs")
		repoDir = filepath.Join(SuiteData.TestDirPath, "repo")
		cacheDir = filepath.Join(SuiteData.TestDirPath, "worktree-cache")
		gitInitRepo(ctx, repoDir)
		writeLFSTestFile(repoDir, "assets/a.bin", "real object contents\x00\xff", true)
		writeLFSTestFile(repoDir, "assets/excluded.bin", "unavailable excluded object", false)
		writeLFSTestFile(repoDir, "outside.bin", "unavailable outside object", false)
		gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitattributes"), []byte("*.bin filter=lfs diff=lfs merge=lfs -text\n"), 0o644)).To(gomega.Succeed())
		gitSucceed(ctx, repoDir, "add", ".")
		gitCommitSucceed(ctx, repoDir, "-m", "LFS pointers")
		gitSucceed(ctx, repoDir, "remote", "add", "origin", "file://"+repoDir)
		commit = gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
	})

	ginkgo.DescribeTable("materializes only selected files independently of host LFS settings",
		func(ctx ginkgo.SpecContext, scope, checksum string, include []string, limited bool) {
			if limited {
				global := filepath.Join(SuiteData.TestDirPath, "host.gitconfig")
				gomega.Expect(os.WriteFile(global, []byte("[lfs]\nfetchinclude = absent\nfetchexclude = *\n"), 0o644)).To(gomega.Succeed())
				setEnvForSpec("GIT_CONFIG_GLOBAL", global)
				setEnvForSpec("GIT_LFS_SKIP_SMUDGE", "1")
				setEnvForSpec("GIT_LFS_SKIP_DOWNLOAD_ERRORS", "1")
				gitSucceed(ctx, repoDir, "config", "lfs.fetchexclude", "*")
			}
			opts := ArchiveOptions{
				Commit: commit, PathScope: scope, ContentChecksum: checksum,
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
					BasePath: scope, IncludeGlobs: include, ExcludeGlobs: []string{"excluded.bin"},
				}),
			}
			var output bytes.Buffer
			gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, opts)).To(gomega.Succeed())
			files := readTestTar(output.Bytes())
			gomega.Expect(files).To(gomega.HaveLen(1))
			for _, content := range files {
				gomega.Expect(content).To(gomega.Equal("real object contents\x00\xff"))
			}
			gomega.Expect(os.ReadFile(filepath.Join(repoDir, "assets", "a.bin"))).To(gomega.Equal([]byte(lfsTestPointer("real object contents\x00\xff"))))
		},
		ginkgo.Entry("Stapel without git lfs install", "assets", "", []string{"a.bin"}, false),
		ginkgo.Entry("Stapel with host selection settings", "assets", "", []string{"a.bin"}, true),
		ginkgo.Entry("Stapel exclusion without an include mask", "assets", "", nil, true),
		ginkgo.Entry("Dockerfile .dockerignore filtering", "", "context-checksum", []string{"assets/a.bin"}, false),
		ginkgo.Entry("Dockerfile with host selection settings", "", "context-checksum", []string{"assets/a.bin"}, true),
		ginkgo.Entry("file scope with host selection settings", "assets/a.bin", "", nil, true),
	)

	ginkgo.It("fails for an included object unavailable on the server", func(ctx ginkgo.SpecContext) {
		var output bytes.Buffer
		err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
			Commit: commit, PathScope: "outside.bin", PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("outside.bin"))
	})

	ginkgo.DescribeTable("rejects successful smudge output that does not match the pointer",
		func(ctx ginkgo.SpecContext, fakeContent string) {
			if len(fakeContent) == len("real object contents\x00\xff") {
				gomega.Expect(sha256.Sum256([]byte(fakeContent))).NotTo(gomega.Equal(sha256.Sum256([]byte("real object contents\x00\xff"))))
			}
			fakeBin := filepath.Join(SuiteData.TestDirPath, "fake-bin")
			gomega.Expect(os.Mkdir(fakeBin, 0o755)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(fakeBin, "git-lfs"), []byte("#!/bin/sh\nprintf '%s' '"+fakeContent+"'\n"), 0o755)).To(gomega.Succeed())
			setEnvForSpec("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
			var output bytes.Buffer
			err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: commit, PathScope: "assets/a.bin", PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("does not match pointer")))
		},
		ginkgo.Entry("SHA-256 mismatch with equal size", strings.Repeat("x", len("real object contents\x00\xff"))),
		ginkgo.Entry("size mismatch", "short"),
	)

	ginkgo.It("downloads public HTTP objects through the standard LFS batch API", func(ctx ginkgo.SpecContext) {
		content := "HTTP object contents\x00"
		oid := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		var requests atomic.Int32
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			requests.Add(1)
			gomega.Expect(r.Header.Get("Authorization")).To(gomega.BeEmpty())
			if r.Method == http.MethodPost && r.URL.Path == "/repo.git/info/lfs/objects/batch" {
				w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
				gomega.Expect(json.NewEncoder(w).Encode(map[string]any{"transfer": "basic", "objects": []any{map[string]any{
					"oid": oid, "size": len(content), "actions": map[string]any{"download": map[string]any{"href": server.URL + "/object"}},
				}}})).To(gomega.Succeed())
				return
			}
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			gomega.Expect(r.URL.Path).To(gomega.Equal("/object"))
			_, err := w.Write([]byte(content))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}))
		defer server.Close()
		writeLFSTestFile(repoDir, "assets/http.bin", content, false)
		gitSucceed(ctx, repoDir, "add", "assets/http.bin")
		gitCommitSucceed(ctx, repoDir, "-m", "HTTP pointer")
		commit = gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
		gitSucceed(ctx, repoDir, "remote", "set-url", "origin", server.URL+"/repo.git")
		var output bytes.Buffer
		gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
			Commit: commit, PathScope: "assets/http.bin",
			FileRenames: map[string]string{"assets/http.bin": "http.bin"},
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})).To(gomega.Succeed())
		gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{"http.bin": content}))
		gomega.Expect(requests.Load()).To(gomega.Equal(int32(2)))
	})

	ginkgo.It("requires git-lfs only for selected LFS pointers", func(ctx ginkgo.SpecContext) {
		gitPath, err := exec.LookPath("git")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gitOnlyPath := filepath.Join(SuiteData.TestDirPath, "git-only")
		gomega.Expect(os.Mkdir(gitOnlyPath, 0o755)).To(gomega.Succeed())
		gomega.Expect(os.Symlink(gitPath, filepath.Join(gitOnlyPath, "git"))).To(gomega.Succeed())
		setEnvForSpec("PATH", gitOnlyPath)
		var output bytes.Buffer
		opts := ArchiveOptions{Commit: commit, PathScope: ".gitattributes", PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{})}
		gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, opts)).To(gomega.Succeed())
		gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": "*.bin filter=lfs diff=lfs merge=lfs -text\n"}))
		opts.PathScope = "assets/a.bin"
		output.Reset()
		err = Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, opts)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("git-lfs is required")))
	})

	ginkgo.It("rejects LFS pointers inside submodules", func(ctx ginkgo.SpecContext) {
		gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())
		gitSucceed(ctx, repoDir, "config", "protocol.file.allow", "always")
		gitAddSubmoduleSucceed(ctx, repoDir, repoDir, "sub")
		gitCommitSucceed(ctx, repoDir, "-m", "LFS submodule")
		commit = gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
		setEnvForSpec("GIT_CONFIG_COUNT", "1")
		setEnvForSpec("GIT_CONFIG_KEY_0", "protocol.file.allow")
		setEnvForSpec("GIT_CONFIG_VALUE_0", "always")
		var output bytes.Buffer
		err := ArchiveWithSubmodules(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
			Commit: commit, PathScope: "sub/assets/a.bin", PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("in submodule \"sub\" is not supported")))
	})

	ginkgo.It("reads pointers from an alternate object store in a shared clone", func(ctx ginkgo.SpecContext) {
		cloneDir := filepath.Join(SuiteData.TestDirPath, "shared-clone")
		gitSucceed(ctx, SuiteData.TestDirPath, "clone", "--shared", repoDir, cloneDir)
		var output bytes.Buffer
		gomega.Expect(Archive(ctx, &output, filepath.Join(cloneDir, ".git"), cacheDir, ArchiveOptions{
			Commit: commit, PathScope: "assets/a.bin",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})).To(gomega.Succeed())
		gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": "real object contents\x00\xff"}))
	})
})
