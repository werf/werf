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
