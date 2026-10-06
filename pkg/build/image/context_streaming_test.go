package image

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
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

	ginkgo.It("pins the cached context next to the archive cache instead of the temp dir", func(ctx ginkgo.SpecContext) {
		projectDir := newProjectRepo(ctx, contextStreamingProjectFiles())

		archive := dockerfileContextArchives(ctx, projectDir, "one")[0]

		gomega.Expect(archive.path).To(gomega.HavePrefix(werf.GetServiceDir()+string(os.PathSeparator)),
			"the pin must share a filesystem with the archive cache, otherwise os.Link fails with EXDEV")
		gomega.Expect(archive.path).NotTo(gomega.HavePrefix(werf.GetTmpDir() + string(os.PathSeparator)))
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
		originalPinDirFunc := createContextPinDir
		createContextPinDir = func(context.Context) (string, error) { return os.MkdirTemp(dir, "pin") }
		ginkgo.DeferCleanup(func() { createContextPinDir = originalPinDirFunc })
		archive := NewBuildContextArchive(giterminismManagerOf(ctx, projectDir), ginkgo.GinkgoT().TempDir())
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
