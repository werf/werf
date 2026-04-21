package container_backend

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	copyrec "github.com/werf/copy-recurse"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("BuildahBackend data archives", func() {
	BeforeEach(func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())
	})

	It("extractTarWithChown applies ownership to extracted archive entries", func() {
		dstDir := GinkgoT().TempDir()
		preExistingDir := filepath.Join(dstDir, "preexisting")
		Expect(os.MkdirAll(preExistingDir, 0o755)).To(Succeed())

		uid := uint32(1001)
		gid := uint32(1001)

		archiveData := newTestTarArchive(map[string]string{"newfile.txt": "content"})
		err := extractTarWithChown(archiveData, dstDir, &uid, &gid)
		if runtime.GOOS != "linux" {
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "chown")).To(BeTrue())
			return
		}

		Expect(err).ToNot(HaveOccurred())
		assertOwnership(filepath.Join(dstDir, "newfile.txt"), 1001, 1001)
	})

	It("extractTarWithChown works without ownership when uid/gid are nil", func() {
		dstDir := GinkgoT().TempDir()
		archiveData := newTestTarArchive(map[string]string{"file.txt": "data"})

		Expect(extractTarWithChown(archiveData, dstDir, nil, nil)).To(Succeed())

		data, err := os.ReadFile(filepath.Join(dstDir, "file.txt"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal("data"))
	})

	It("applyDataArchives extracts and applies string owner/group", func(ctx SpecContext) {
		var testCtx context.Context = logging.WithLogger(ctx)
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "etc"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "etc", "passwd"), []byte("gitlab:x:1001:1001::/home/gitlab:/bin/sh\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(rootMount, "etc", "group"), []byte("gitlab:x:1001:\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(rootMount, "app"), 0o755)).To(Succeed())

		archiveReader := newTestTarArchive(map[string]string{"README.md": "content"})
		backend := &BuildahBackend{}

		err := backend.applyDataArchives(testCtx, &containerDesc{RootMount: rootMount}, []DataArchiveSpec{{
			Archive: archiveReader,
			Type:    DirectoryArchive,
			To:      "/app",
			Owner:   "gitlab",
			Group:   "gitlab",
		}})
		if runtime.GOOS != "linux" {
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "chown")).To(BeTrue())
			return
		}

		Expect(err).ToNot(HaveOccurred())
		data, err := os.ReadFile(filepath.Join(rootMount, "app", "README.md"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal("content"))
		assertOwnership(filepath.Join(rootMount, "app", "README.md"), 1001, 1001)
	})

	It("applyDataArchives without owner/group does not chown", func(ctx SpecContext) {
		var testCtx context.Context = logging.WithLogger(ctx)
		rootMount := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(rootMount, "app"), 0o755)).To(Succeed())

		archiveReader := newTestTarArchive(map[string]string{"file.txt": "data"})
		backend := &BuildahBackend{}

		Expect(backend.applyDataArchives(testCtx, &containerDesc{RootMount: rootMount}, []DataArchiveSpec{{
			Archive: archiveReader,
			Type:    DirectoryArchive,
			To:      "/app",
		}})).To(Succeed())

		data, err := os.ReadFile(filepath.Join(rootMount, "app", "file.txt"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal("data"))
	})

	It("normalizes dependency import destination for regular file into root directory", func() {
		sourceRoot := GinkgoT().TempDir()
		containerRoot := GinkgoT().TempDir()

		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(sourceRoot, "src"), 0o755))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(sourceRoot, "src", "webhook"), []byte("webhook\n"), 0o644))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(containerRoot, "sentinel"), []byte("keep\n"), 0o644))

		absFrom := filepath.Join(sourceRoot, "src", "webhook")
		absTo, err := normalizeDependencyImportDestination(absFrom, containerRoot)
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), filepath.Join(containerRoot, "webhook"), absTo)

		copyRec, err := copyrec.New(absFrom, absTo, copyrec.Options{})
		require.NoError(GinkgoT(), err)
		require.NoError(GinkgoT(), copyRec.Run(context.Background()))

		info, err := os.Lstat(containerRoot)
		require.NoError(GinkgoT(), err)
		assert.True(GinkgoT(), info.IsDir())

		data, err := os.ReadFile(filepath.Join(containerRoot, "webhook"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "webhook\n", string(data))

		data, err = os.ReadFile(filepath.Join(containerRoot, "sentinel"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "keep\n", string(data))
	})

	It("keeps explicit dependency import file destination unchanged", func() {
		sourceRoot := GinkgoT().TempDir()
		containerRoot := GinkgoT().TempDir()

		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(sourceRoot, "src"), 0o755))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(sourceRoot, "src", "webhook"), []byte("webhook\n"), 0o644))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(containerRoot, "sentinel"), []byte("keep\n"), 0o644))

		absFrom := filepath.Join(sourceRoot, "src", "webhook")
		explicitTarget := filepath.Join(containerRoot, "webhook")
		absTo, err := normalizeDependencyImportDestination(absFrom, explicitTarget)
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), explicitTarget, absTo)

		copyRec, err := copyrec.New(absFrom, absTo, copyrec.Options{})
		require.NoError(GinkgoT(), err)
		require.NoError(GinkgoT(), copyRec.Run(context.Background()))

		info, err := os.Lstat(containerRoot)
		require.NoError(GinkgoT(), err)
		assert.True(GinkgoT(), info.IsDir())

		data, err := os.ReadFile(filepath.Join(containerRoot, "webhook"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "webhook\n", string(data))

		data, err = os.ReadFile(filepath.Join(containerRoot, "sentinel"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "keep\n", string(data))
	})

	It("normalizes dependency import destination for symlink into root directory", func() {
		sourceRoot := GinkgoT().TempDir()
		containerRoot := GinkgoT().TempDir()

		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(sourceRoot, "src"), 0o755))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(sourceRoot, "src", "webhook"), []byte("webhook\n"), 0o644))
		createTestSymlink(filepath.Join(sourceRoot, "src", "webhook"), filepath.Join(sourceRoot, "src", "webhook-link"))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(containerRoot, "sentinel"), []byte("keep\n"), 0o644))

		absFrom := filepath.Join(sourceRoot, "src", "webhook-link")
		absTo, err := normalizeDependencyImportDestination(absFrom, containerRoot)
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), filepath.Join(containerRoot, "webhook-link"), absTo)

		copyRec, err := copyrec.New(absFrom, absTo, copyrec.Options{})
		require.NoError(GinkgoT(), err)
		require.NoError(GinkgoT(), copyRec.Run(context.Background()))

		info, err := os.Lstat(containerRoot)
		require.NoError(GinkgoT(), err)
		assert.True(GinkgoT(), info.IsDir())

		linkTarget, err := os.Readlink(filepath.Join(containerRoot, "webhook-link"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), filepath.Join(sourceRoot, "src", "webhook"), linkTarget)

		data, err := os.ReadFile(filepath.Join(containerRoot, "sentinel"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "keep\n", string(data))
	})

	It("keeps explicit dependency import symlink destination unchanged", func() {
		sourceRoot := GinkgoT().TempDir()
		containerRoot := GinkgoT().TempDir()

		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(sourceRoot, "src"), 0o755))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(sourceRoot, "src", "webhook"), []byte("webhook\n"), 0o644))
		createTestSymlink(filepath.Join(sourceRoot, "src", "webhook"), filepath.Join(sourceRoot, "src", "webhook-link"))
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(containerRoot, "sentinel"), []byte("keep\n"), 0o644))

		absFrom := filepath.Join(sourceRoot, "src", "webhook-link")
		explicitTarget := filepath.Join(containerRoot, "webhook-link")
		absTo, err := normalizeDependencyImportDestination(absFrom, explicitTarget)
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), explicitTarget, absTo)

		copyRec, err := copyrec.New(absFrom, absTo, copyrec.Options{})
		require.NoError(GinkgoT(), err)
		require.NoError(GinkgoT(), copyRec.Run(context.Background()))

		info, err := os.Lstat(containerRoot)
		require.NoError(GinkgoT(), err)
		assert.True(GinkgoT(), info.IsDir())

		linkTarget, err := os.Readlink(filepath.Join(containerRoot, "webhook-link"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), filepath.Join(sourceRoot, "src", "webhook"), linkTarget)

		data, err := os.ReadFile(filepath.Join(containerRoot, "sentinel"))
		require.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), "keep\n", string(data))
	})
})

func assertOwnership(path string, uid, gid uint32) {
	info, err := os.Lstat(path)
	Expect(err).ToNot(HaveOccurred())
	stat, ok := info.Sys().(*syscall.Stat_t)
	Expect(ok).To(BeTrue())
	Expect(stat.Uid).To(Equal(uid))
	Expect(stat.Gid).To(Equal(gid))
}

func createTestSymlink(oldname, newname string) {
	if runtime.GOOS == "windows" {
		Skip("skip on windows")
	}

	require.NoError(GinkgoT(), os.Symlink(oldname, newname))
}

func newTestTarArchive(files map[string]string) io.ReadCloser {
	var buf bytes.Buffer
	writer := tar.NewWriter(&buf)

	for name, content := range files {
		Expect(writer.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		})).To(Succeed())
		_, err := writer.Write([]byte(content))
		Expect(err).ToNot(HaveOccurred())
	}

	Expect(writer.Close()).To(Succeed())

	return io.NopCloser(bytes.NewReader(buf.Bytes()))
}
