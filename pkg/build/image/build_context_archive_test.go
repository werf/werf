package image

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("build context streaming", func() {
	ginkgo.It("does not materialize a full context copy per dockerfile image", func(ctx ginkgo.SpecContext) {
		projectDir := newProjectRepo(ctx, contextStreamingProjectFiles())
		tmpDir := werf.GetTmpDir()

		bytesBefore, err := uniqueFileBytes(tmpDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(dockerfileContextArchives(ctx, projectDir, "one", "two", "three")).To(gomega.HaveLen(3))

		gomega.Expect(filesMatching(tmpDir, "werf-*-context-*")).To(gomega.BeEmpty(), "per-image context archive copies must not be created")

		bytesAfter, err := uniqueFileBytes(tmpDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(bytesAfter-bytesBefore).To(gomega.BeNumerically("<", 2*contextStreamingBlobSize),
			"three contexts over the same commit must not add a full context copy each")
	})

	ginkgo.It("falls back to one materialized overlay when hard links are unavailable", func(ctx ginkgo.SpecContext) {
		dir, err := os.MkdirTemp("/dev/shm", "werf-context-test-")
		if os.IsNotExist(err) || os.IsPermission(err) {
			ginkgo.Skip("requires a writable /dev/shm on a separate filesystem")
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(os.RemoveAll, dir)
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{"one.Dockerfile": "FROM scratch\n"}))
		if err := os.Link(filepath.Join(projectDir, "one.Dockerfile"), filepath.Join(dir, "probe")); err == nil {
			ginkgo.Skip("/dev/shm supports hard links from the project filesystem")
		}
		archive := NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), dir)
		gomega.Expect(archive.Create(ctx, container_backend.BuildContextArchiveCreateOptions{
			DockerfileRelToContextPath: "one.Dockerfile",
		})).To(gomega.Succeed())
		entries := openedContextEntries(ctx, archive)
		gomega.Expect(entries).To(gomega.Equal(tarFileEntries(archive.path)))
		var modes []int64
		for _, entry := range entries {
			if entry.Name == "one.Dockerfile" {
				modes = append(modes, entry.Mode)
			}
		}
		gomega.Expect(modes).To(gomega.Equal([]int64{0o100644, 0o600}))
	})

	ginkgo.DescribeTable("streams exactly what the materialized context archive contained",
		func(ctx ginkgo.SpecContext, files map[string]string, setup func(ctx context.Context, projectDir string), imageName, contextSubDir, dockerfileEntry, dockerfileContents string) {
			projectDir := newProjectRepo(ctx, files)
			if setup != nil {
				setup(ctx, projectDir)
			}

			archive := dockerfileContextArchives(ctx, projectDir, imageName)[0]

			reference := tarFileEntries(materializedContextArchive(ctx, archive, projectDir, contextSubDir))
			gomega.Expect(len(reference)).To(gomega.BeNumerically(">=", 2))
			gomega.Expect(reference[len(reference)-1].Name).To(gomega.Equal(dockerfileEntry), "the dockerfile overlay entry must be appended last")
			gomega.Expect(reference[len(reference)-1].Mode).To(gomega.Equal(int64(0o600)))

			entries := openedContextEntries(ctx, archive)
			gomega.Expect(entries).To(gomega.Equal(reference))
			gomega.Expect(lastEntryContents(entries)).To(gomega.HaveKeyWithValue(dockerfileEntry, dockerfileContents))
			if _, err := os.Lstat(filepath.Join(projectDir, "link.txt")); err == nil {
				gomega.Expect(lastEntryContents(entries)).To(gomega.HaveKeyWithValue("data.txt", "payload\n"))
				gomega.Expect(entries).To(gomega.ContainElement(gomega.SatisfyAll(
					gomega.HaveField("Name", "link.txt"),
					gomega.HaveField("Typeflag", uint8(tar.TypeSymlink)),
					gomega.HaveField("Linkname", "data.txt"),
				)))
				var modes []int64
				for _, entry := range entries {
					if entry.Name == dockerfileEntry {
						modes = append(modes, entry.Mode)
					}
				}
				gomega.Expect(modes).To(gomega.Equal([]int64{0o100755, 0o600}))
			} else {
				gomega.Expect(err).To(gomega.MatchError(os.ErrNotExist))
			}
		},
		ginkgo.Entry("dockerfile at the context root", dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{"one.Dockerfile": "FROM scratch\nCOPY data.txt /\n", "data.txt": "payload\n"}),
			nil, "one", "", "one.Dockerfile", "FROM scratch\nCOPY data.txt /\n"),
		ginkgo.Entry("nested context dir", dockerfileProjectFiles("\nimage: nested\ncontext: sub\ndockerfile: deep/nested.Dockerfile\n",
			map[string]string{
				"sub/deep/nested.Dockerfile": "FROM scratch\nCOPY inner.txt /\n",
				"sub/inner.txt":              "inner\n",
				"outside.txt":                "outside\n",
			}), nil, "nested", "sub", "deep/nested.Dockerfile", "FROM scratch\nCOPY inner.txt /\n"),
		ginkgo.Entry("executable dockerfile and tracked symlink", dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{"one.Dockerfile": "FROM scratch\nCOPY data.txt /\n", "data.txt": "payload\n"}),
			func(ctx context.Context, projectDir string) {
				utils.RunSucceedCommand(ctx, projectDir, "chmod", "+x", "one.Dockerfile")
				utils.RunSucceedCommand(ctx, projectDir, "ln", "-s", "data.txt", "link.txt")
				commitFiles(ctx, projectDir, nil)
			}, "one", "", "one.Dockerfile", "FROM scratch\nCOPY data.txt /\n"),
		ginkgo.Entry("allowed uncommitted dockerfile", dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{
				"werf-giterminism.yaml": giterminismDockerfileConfig(map[string][]string{"allowUncommitted": {"one.Dockerfile"}}),
				"one.Dockerfile":        "FROM scratch\nCOPY data.txt /\n",
				"data.txt":              "payload\n",
			}),
			func(ctx context.Context, projectDir string) {
				utils.WriteFile(filepath.Join(projectDir, "one.Dockerfile"), []byte("FROM scratch\n# uncommitted\n"))
			}, "one", "", "one.Dockerfile", "FROM scratch\n# uncommitted\n"),
		ginkgo.Entry("allowed untracked dockerfile", dockerfileProjectFiles("\nimage: one\ndockerfile: untracked.Dockerfile\n",
			map[string]string{
				"werf-giterminism.yaml": giterminismDockerfileConfig(map[string][]string{"allowUncommitted": {"untracked.Dockerfile"}}),
				"data.txt":              "payload\n",
			}),
			func(ctx context.Context, projectDir string) {
				utils.WriteFile(filepath.Join(projectDir, "untracked.Dockerfile"), []byte("FROM scratch\n# untracked\n"))
			}, "one", "", "untracked.Dockerfile", "FROM scratch\n# untracked\n"),
	)

	ginkgo.It("keeps the dockerfile overlay winning over contextAddFiles, which win over git", func(ctx ginkgo.SpecContext) {
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\ncontextAddFiles:\n- added.txt\n- one.Dockerfile\n",
			map[string]string{
				"werf-giterminism.yaml": giterminismDockerfileConfig(map[string][]string{
					"allowContextAddFiles": {"added.txt", "one.Dockerfile"},
					"allowUncommitted":     {"one.Dockerfile"},
				}),
				"one.Dockerfile": "FROM scratch\nCOPY added.txt /\n",
				"added.txt":      "committed\n",
			}))
		utils.WriteFile(filepath.Join(projectDir, "added.txt"), []byte("from worktree\n"))
		utils.WriteFile(filepath.Join(projectDir, "one.Dockerfile"), []byte("FROM scratch\n# from worktree\n"))
		utils.RunSucceedCommand(ctx, projectDir, "chmod", "+x", "one.Dockerfile")

		archive := dockerfileContextArchives(ctx, projectDir, "one")[0]
		entries := openedContextEntries(ctx, archive)

		gomega.Expect(entries).To(gomega.Equal(tarFileEntries(archive.path)), "the contextAddFiles fallback must stream its materialized archive as is")
		gomega.Expect(lastEntryContents(entries)).To(gomega.SatisfyAll(
			gomega.HaveKeyWithValue("added.txt", "from worktree\n"),
			gomega.HaveKeyWithValue("one.Dockerfile", "FROM scratch\n# from worktree\n"),
		))

		modes := map[string][]int64{}
		for _, entry := range entries {
			modes[entry.Name] = append(modes[entry.Name], entry.Mode)
		}
		gomega.Expect(modes["one.Dockerfile"]).To(gomega.Equal([]int64{0o755, 0o600}),
			"the contextAddFiles copy must be followed, and so overridden, by the dockerfile overlay")
	})

	ginkgo.It("extracts the same tree the materialized archive extracts to", func(ctx ginkgo.SpecContext) {
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{
				"one.Dockerfile":        "FROM scratch\nCOPY data.txt /\n",
				"data.txt":              "payload\n",
				"werf-giterminism.yaml": giterminismDockerfileConfig(map[string][]string{"allowUncommitted": {"one.Dockerfile"}}),
			}))
		utils.WriteFile(filepath.Join(projectDir, "one.Dockerfile"), []byte("FROM scratch\nCOPY data.txt /changed\n"))
		archive := dockerfileContextArchives(ctx, projectDir, "one")[0]

		referenceFile, err := os.Open(materializedContextArchive(ctx, archive, projectDir, ""))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(referenceFile.Close)

		referenceDir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(util.ExtractTar(referenceFile, referenceDir, util.ExtractTarOptions{})).To(gomega.Succeed())

		extractedDir, err := archive.ExtractOrGetExtractedDir(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(dirEntries(extractedDir)).To(gomega.Equal(dirEntries(referenceDir)))
		gomega.Expect(dirEntries(extractedDir)).To(gomega.HaveKey("one.Dockerfile"))
	})

	ginkgo.It("serves repeated, concurrent and cache-evicted readers", func(ctx ginkgo.SpecContext) {
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{"one.Dockerfile": "FROM scratch\n", "blob.bin": strings.Repeat("y", contextStreamingBlobSize)}))
		archive := dockerfileContextArchives(ctx, projectDir, "one")[0]

		expected := openedContextEntries(ctx, archive)
		gomega.Expect(openedContextEntries(ctx, archive)).To(gomega.Equal(expected), "a second Open must return the same stream")

		abandoned, err := archive.Open(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(abandoned.Close()).To(gomega.Succeed())

		gomega.Expect(os.RemoveAll(filepath.Join(werf.GetLocalCacheDir(), "git_archives"))).To(gomega.Succeed())

		var wg sync.WaitGroup
		results := make([][]tarEntry, 3)
		for i := range results {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer ginkgo.GinkgoRecover()
				results[i] = openedContextEntries(ctx, archive)
			}()
		}
		wg.Wait()

		for _, result := range results {
			gomega.Expect(result).To(gomega.Equal(expected))
		}
	})

	ginkgo.It("fails the stream on cancellation, a missing archive and a corrupt archive", func(ctx ginkgo.SpecContext) {
		projectDir := newProjectRepo(ctx, dockerfileProjectFiles("\nimage: one\ndockerfile: one.Dockerfile\n",
			map[string]string{"one.Dockerfile": "FROM scratch\n", "blob.bin": strings.Repeat("z", contextStreamingBlobSize)}))
		archive := dockerfileContextArchives(ctx, projectDir, "one")[0]

		canceledCtx, cancel := context.WithCancel(ctx)
		reader, err := archive.Open(canceledCtx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		cancel()
		gomega.Eventually(func() error {
			_, err := reader.Read(make([]byte, 1))
			return err
		}).Should(gomega.MatchError(context.Canceled), "cancellation after Open must reach the consumer")
		gomega.Expect(reader.Close()).To(gomega.Succeed())

		_, err = archive.Open(canceledCtx)
		gomega.Expect(err).To(gomega.MatchError(context.Canceled))

		corrupt := filepath.Join(ginkgo.GinkgoT().TempDir(), "corrupt.tar")
		gomega.Expect(os.WriteFile(corrupt, []byte("this is not a tar archive"), 0o644)).To(gomega.Succeed())
		corruptArchive := &BuildContextArchive{path: corrupt, contextAddFilesFromMem: archive.contextAddFilesFromMem, extractionRootTmpDir: ginkgo.GinkgoT().TempDir()}
		corruptReader, err := corruptArchive.Open(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = io.Copy(io.Discard, corruptReader)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(corruptReader.Close()).To(gomega.Succeed())
		for range 2 {
			_, err := corruptArchive.ExtractOrGetExtractedDir(ctx)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(corruptArchive.extractionDir).To(gomega.BeEmpty())
			gomega.Expect(os.ReadDir(corruptArchive.extractionRootTmpDir)).To(gomega.BeEmpty())
		}

		gomega.Expect(os.Remove(archive.path)).To(gomega.Succeed())
		_, err = archive.Open(ctx)
		gomega.Expect(err).To(gomega.MatchError(os.ErrNotExist))
	})
})

var _ = ginkgo.Describe("build context archive content caching", func() {
	ginkgo.DescribeTable("recreates the cached context archive only when the context contents change",
		func(ctx ginkgo.SpecContext, mutate func(ctx context.Context, projectDir string), reused bool, verify func(entries []tarEntry)) {
			projectDir := newContentCachingRepo(ctx)
			if reused {
				requireGitAttributeSource(ctx, projectDir)
			}

			beforePath, before := cachedArchive(ctx, projectDir)
			beforeEntries := tarFileEntries(beforePath)

			mutate(ctx, projectDir)

			afterPath, after := cachedArchive(ctx, projectDir)
			afterEntries := tarFileEntries(afterPath)

			gomega.Expect(os.SameFile(before, after)).To(gomega.Equal(reused),
				"the cached archive must be reused only when the context contents are unchanged")
			if reused {
				gomega.Expect(afterEntries).To(gomega.Equal(beforeEntries))
			} else {
				gomega.Expect(afterEntries).NotTo(gomega.Equal(beforeEntries))
			}
			verify(afterEntries)
		},
		ginkgo.Entry("a commit outside the context", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
		}, true, func(entries []tarEntry) {
			gomega.Expect(lastEntryContents(entries)).NotTo(gomega.HaveKey("outside.txt"))
		}),
		ginkgo.Entry("a commit of a dockerignored file", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{"app/ignored.txt": "changed\n"})
		}, true, func(entries []tarEntry) {
			gomega.Expect(lastEntryContents(entries)).NotTo(gomega.HaveKey("ignored.txt"))
		}),
		ginkgo.Entry("equivalent ignore rules", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{"app/.dockerignore": ".dockerignore\nignored.txt\nmissing.txt\n"})
		}, true, func(entries []tarEntry) {
			gomega.Expect(lastEntryContents(entries)).NotTo(gomega.HaveKey(".dockerignore"))
		}),
		ginkgo.Entry("a commit changing included contents", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{"app/included.txt": "changed\n"})
		}, false, func(entries []tarEntry) {
			gomega.Expect(lastEntryContents(entries)).To(gomega.HaveKeyWithValue("included.txt", "changed\n"))
		}),
		ginkgo.Entry("a commit changing an included file mode", func(ctx context.Context, projectDir string) {
			utils.RunSucceedCommand(ctx, projectDir, "chmod", "+x", "app/included.txt")
			commitFiles(ctx, projectDir, nil)
		}, false, func(entries []tarEntry) {
			gomega.Expect(entryNamed(entries, "included.txt").Mode & 0o777).To(gomega.Equal(int64(0o755)))
		}),
		ginkgo.Entry("a commit renaming an included file", func(ctx context.Context, projectDir string) {
			utils.RunSucceedCommand(ctx, projectDir, "git", "mv", "app/included.txt", "app/renamed.txt")
			commitFiles(ctx, projectDir, nil)
		}, false, func(entries []tarEntry) {
			gomega.Expect(lastEntryContents(entries)).To(gomega.SatisfyAll(
				gomega.HaveKeyWithValue("renamed.txt", "included\n"),
				gomega.Not(gomega.HaveKey("included.txt")),
			))
		}),
		ginkgo.Entry("a commit repointing a tracked symlink", func(ctx context.Context, projectDir string) {
			utils.RunSucceedCommand(ctx, projectDir, "ln", "-sf", "ignored.txt", "app/link.txt")
			commitFiles(ctx, projectDir, nil)
		}, false, func(entries []tarEntry) {
			gomega.Expect(entryNamed(entries, "link.txt").Linkname).To(gomega.Equal("ignored.txt"))
		}),
	)

	ginkgo.It("returns to the original cached archive once the original contents are restored", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		requireGitAttributeSource(ctx, projectDir)

		originalPath, original := cachedArchive(ctx, projectDir)
		originalEntries := tarFileEntries(originalPath)

		commitFiles(ctx, projectDir, map[string]string{"app/included.txt": "changed\n"})
		_, changed := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(original, changed)).To(gomega.BeFalse())

		commitFiles(ctx, projectDir, map[string]string{"app/included.txt": "included\n"})
		restoredPath, restored := cachedArchive(ctx, projectDir)

		gomega.Expect(os.SameFile(original, restored)).To(gomega.BeTrue(), "restored contents must hit the original cached archive")
		gomega.Expect(tarFileEntries(restoredPath)).To(gomega.Equal(originalEntries))
	})

	ginkgo.It("keeps the cached archive while an allowed uncommitted dockerfile changes the stream", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		commitFiles(ctx, projectDir, map[string]string{
			"werf-giterminism.yaml": giterminismDockerfileConfig(map[string][]string{"allowUncommitted": {"app/Dockerfile"}}),
		})

		_, before := cachedArchive(ctx, projectDir)

		utils.WriteFile(filepath.Join(projectDir, "app/Dockerfile"), []byte("FROM scratch\n# uncommitted\n"))
		archive := dockerfileContextArchives(ctx, projectDir, projectImageName)[0]
		after, err := os.Stat(archive.path)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(os.SameFile(before, after)).To(gomega.BeTrue(), "an uncommitted dockerfile must not invalidate the cached context archive")
		gomega.Expect(lastEntryContents(openedContextEntries(ctx, archive))).To(gomega.HaveKeyWithValue("Dockerfile", "FROM scratch\n# uncommitted\n"))
	})
})

func newBuildContextArchive(t *testing.T, dirName string) *BuildContextArchive {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, dirName), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, dirName, "x"), []byte("data"), 0o644))

	return &BuildContextArchive{path: filepath.Join(root, "context.tar"), extractionDir: root}
}

func TestCalculateGlobsChecksumMatchedPaths(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name  string
		opts  container_backend.CalculateGlobsChecksumOptions
		equal bool
	}{
		{"contents only", container_backend.CalculateGlobsChecksumOptions{}, true},
		{"matched paths included", container_backend.CalculateGlobsChecksumOptions{IncludeMatchedPaths: true}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checksumAAA, err := newBuildContextArchive(t, "aaa").CalculateGlobsChecksum(ctx, []string{"./*"}, tt.opts)
			require.NoError(t, err)

			checksumCCC, err := newBuildContextArchive(t, "ccc").CalculateGlobsChecksum(ctx, []string{"./*"}, tt.opts)
			require.NoError(t, err)

			if tt.equal {
				require.Equal(t, checksumAAA, checksumCCC)
			} else {
				require.NotEqual(t, checksumAAA, checksumCCC)
			}
		})
	}
}

var _ = ginkgo.Describe("build context archive checkout safety", func() {
	ginkgo.It("checks committed attributes instead of dirty worktree attributes", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt text eol=crlf\n"})
		_, before := cachedArchive(ctx, projectDir)
		utils.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte("# no conversions\n"))
		path, after := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(before, after)).To(gomega.BeTrue())
		gomega.Expect(lastEntryContents(tarFileEntries(path))).To(gomega.HaveKeyWithValue("included.txt", "included\r\n"))
	})

	ginkgo.It("does not cache stale checkout bytes after conversion attributes are removed", func(ctx ginkgo.SpecContext) {
		projectDir, err := filepath.EvalSymlinks(newContentCachingRepo(ctx))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		requireGitAttributeSource(ctx, projectDir)
		commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt text eol=crlf\n"})
		convertedPath, _ := cachedArchive(ctx, projectDir)
		gomega.Expect(lastEntryContents(tarFileEntries(convertedPath))).To(gomega.HaveKeyWithValue("included.txt", "included\r\n"))

		utils.RunSucceedCommand(ctx, projectDir, "git", "rm", ".gitattributes")
		commitFiles(ctx, projectDir, nil)
		path, info := cachedArchive(ctx, projectDir)
		gomega.Expect(lastEntryContents(tarFileEntries(path))).To(gomega.HaveKeyWithValue("included.txt", "included\n"))

		commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
		reusedPath, reused := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(info, reused)).To(gomega.BeTrue())
		gomega.Expect(lastEntryContents(tarFileEntries(reusedPath))).To(gomega.HaveKeyWithValue("included.txt", "included\n"))
	})

	ginkgo.DescribeTable("falls back when attribute inspection is unreliable",
		func(ctx ginkgo.SpecContext, response string) {
			projectDir := newContentCachingRepo(ctx)
			interceptGitCheckAttributes(response)
			_, before := cachedArchive(ctx, projectDir)
			commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
			_, after := cachedArchive(ctx, projectDir)
			gomega.Expect(os.SameFile(before, after)).To(gomega.BeFalse())
		},
		ginkgo.Entry("unsupported Git option", "exit 129"),
		ginkgo.Entry("warnings", "printf warning >&2; exit 0"),
		ginkgo.Entry("truncated output", "printf invalid; exit 0"),
		ginkgo.Entry("unexpected path", "printf 'not-selected\\000diff\\000set\\000'; exit 0"),
	)

	ginkgo.DescribeTable("falls back for worktree-dependent configuration",
		func(ctx ginkgo.SpecContext, key, value string) {
			projectDir := newContentCachingRepo(ctx)
			if key == "extensions.worktreeConfig" {
				ginkgo.GinkgoT().Setenv("GIT_CONFIG_COUNT", "1")
				ginkgo.GinkgoT().Setenv("GIT_CONFIG_KEY_0", key)
				ginkgo.GinkgoT().Setenv("GIT_CONFIG_VALUE_0", value)
			} else {
				utils.RunSucceedCommand(ctx, projectDir, "git", "config", "--local", key, value)
			}
			_, before := cachedArchive(ctx, projectDir)
			commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
			_, after := cachedArchive(ctx, projectDir)
			gomega.Expect(os.SameFile(before, after)).To(gomega.BeFalse())
		},
		ginkgo.Entry("conditional include", "includeIf.gitdir:never/.path", "/nonexistent"),
		ginkgo.Entry("worktree configuration", "extensions.worktreeConfig", "true"),
		ginkgo.Entry("symlink configuration", "core.symlinks", "true"),
		ginkgo.Entry("attribute tree override", "attr.tree", "HEAD"),
	)

	ginkgo.It("falls back when the attribute source is overridden", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		ginkgo.GinkgoT().Setenv("GIT_ATTR_SOURCE", utils.GetHeadCommit(ctx, projectDir))
		_, before := cachedArchive(ctx, projectDir)
		commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
		_, after := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(before, after)).To(gomega.BeFalse())
	})

	ginkgo.It("falls back when the repository declares submodules", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		commitFiles(ctx, projectDir, map[string]string{".gitmodules": ""})
		_, before := cachedArchive(ctx, projectDir)
		commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
		_, after := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(before, after)).To(gomega.BeFalse())
	})

	ginkgo.DescribeTable("reuses the content-keyed archive across an outside commit only when no selected file can undergo a checkout conversion",
		func(ctx ginkgo.SpecContext, setup func(ctx context.Context, projectDir string), reused bool) {
			projectDir := newContentCachingRepo(ctx)
			if reused {
				requireGitAttributeSource(ctx, projectDir)
			}
			setup(ctx, projectDir)

			_, before := cachedArchive(ctx, projectDir)
			commitFiles(ctx, projectDir, map[string]string{"outside.txt": "changed\n"})
			_, after := cachedArchive(ctx, projectDir)

			gomega.Expect(os.SameFile(before, after)).To(gomega.Equal(reused),
				"a selected file that may be converted on checkout must fall back to the commit-keyed archive")
		},
		ginkgo.Entry("a conversion attribute matching only a dockerignored file", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/ignored.txt text eol=crlf\n"})
		}, true),
		ginkgo.Entry("a root .gitattributes converting an included file", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/*.txt text eol=crlf\n"})
		}, false),
		ginkgo.Entry("text=auto over the whole tree", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "* text=auto\n"})
		}, false),
		ginkgo.Entry("the legacy crlf attribute", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt crlf\n"})
		}, false),
		ginkgo.Entry("ident expansion", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt ident\n"})
		}, false),
		ginkgo.Entry("working-tree-encoding", func(ctx context.Context, projectDir string) {
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt working-tree-encoding=UTF-8\n"})
		}, false),
		ginkgo.Entry("a smudge filter", func(ctx context.Context, projectDir string) {
			utils.RunSucceedCommand(ctx, projectDir, "git", "config", "filter.harmless.smudge", "cat")
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt filter=harmless\n"})
		}, false),
		// git check-attr reports unset and unspecified attributes with the same words a filter may be
		// named after, so a filter literally named "unspecified" must not read as "no filter".
		ginkgo.Entry("a filter named like a check-attr marker", func(ctx context.Context, projectDir string) {
			utils.RunSucceedCommand(ctx, projectDir, "git", "config", "filter.unspecified.smudge", "cat")
			commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/included.txt filter=unspecified\n"})
		}, false),
		ginkgo.Entry("an effective .git/info/attributes", func(ctx context.Context, projectDir string) {
			utils.WriteFile(filepath.Join(projectDir, ".git/info/attributes"), []byte("app/*.txt text eol=crlf\n"))
		}, false),
		ginkgo.Entry("an effective core.attributesFile", func(ctx context.Context, projectDir string) {
			attributesFile := filepath.Join(projectDir, "outside-attributes")
			utils.WriteFile(attributesFile, []byte("*.txt text eol=crlf\n"))
			utils.RunSucceedCommand(ctx, projectDir, "git", "config", "core.attributesFile", attributesFile)
		}, false),
	)

	ginkgo.It("re-materializes the context when a root .gitattributes flips the line endings of an included file", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/*.txt text eol=crlf\n"})

		beforePath, before := cachedArchive(ctx, projectDir)
		gomega.Expect(lastEntryContents(tarFileEntries(beforePath))).To(gomega.HaveKeyWithValue("included.txt", "included\r\n"))

		commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/*.txt text eol=lf\n"})
		// Force a fresh checkout so byte assertions isolate archive reuse from worktree normalization.
		gomega.Expect(os.RemoveAll(git_repo.GetWorkTreeCacheDir())).To(gomega.Succeed())

		afterPath, after := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(before, after)).To(gomega.BeFalse(),
			"flipping eol outside the context must not hit the archive cached for the previous line endings")
		gomega.Expect(lastEntryContents(tarFileEntries(afterPath))).To(gomega.HaveKeyWithValue("included.txt", "included\n"))
	})

	ginkgo.It("resumes content reuse only once the attributes are safe again", func(ctx ginkgo.SpecContext) {
		projectDir := newContentCachingRepo(ctx)
		requireGitAttributeSource(ctx, projectDir)
		_, safe := cachedArchive(ctx, projectDir)

		commitFiles(ctx, projectDir, map[string]string{".gitattributes": "app/*.txt text eol=crlf\n"})
		_, unsafe := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(safe, unsafe)).To(gomega.BeFalse(), "the safe archive must not be reused under a conversion attribute")

		commitFiles(ctx, projectDir, map[string]string{".gitattributes": "# no conversions\n"})
		_, restored := cachedArchive(ctx, projectDir)
		gomega.Expect(os.SameFile(safe, restored)).To(gomega.BeTrue(), "removing the conversion attribute must return to content reuse")
	})
})
