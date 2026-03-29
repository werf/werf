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
})

func assertOwnership(path string, uid, gid uint32) {
	info, err := os.Lstat(path)
	Expect(err).ToNot(HaveOccurred())
	stat, ok := info.Sys().(*syscall.Stat_t)
	Expect(ok).To(BeTrue())
	Expect(stat.Uid).To(Equal(uid))
	Expect(stat.Gid).To(Equal(gid))
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
