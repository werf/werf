package container_backend

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

var _ = Describe("BuildahBackend dependency import destination resolution", func() {
	It("resolves a directory source into a relative symlink destination under the root mount", func() {
		root := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(root, "usr", "bin"), 0o755))
		require.NoError(GinkgoT(), os.Symlink("usr/bin", filepath.Join(root, "bin")))

		src := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(src, "myapp"), []byte("hello"), 0o755))

		dest, err := resolveContainerRootPath(root, "bin")
		Expect(err).ToNot(HaveOccurred())
		absTo, err := normalizeDependencyImportDestination(src, dest)
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(filepath.Join(root, "usr", "bin")))
	})

	It("resolves a file source into a relative symlink destination under the root mount", func() {
		root := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(root, "usr", "bin"), 0o755))
		require.NoError(GinkgoT(), os.Symlink("usr/bin", filepath.Join(root, "bin")))

		srcDir := GinkgoT().TempDir()
		srcFile := filepath.Join(srcDir, "myapp")
		require.NoError(GinkgoT(), os.WriteFile(srcFile, []byte("hello"), 0o755))

		dest, err := resolveContainerRootPath(root, "bin")
		Expect(err).ToNot(HaveOccurred())
		absTo, err := normalizeDependencyImportDestination(srcFile, dest)
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(filepath.Join(root, "usr", "bin", "myapp")))
	})

	It("anchors an absolute symlink destination target under the root mount", func() {
		root := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(root, "usr", "bin"), 0o755))
		require.NoError(GinkgoT(), os.Symlink("/usr/bin", filepath.Join(root, "bin")))

		src := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(src, "myapp"), []byte("hello"), 0o755))

		dest, err := resolveContainerRootPath(root, "bin")
		Expect(err).ToNot(HaveOccurred())
		absTo, err := normalizeDependencyImportDestination(src, dest)
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(filepath.Join(root, "usr", "bin")))
	})

	It("leaves a non-symlink directory destination unchanged", func() {
		root := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(root, "opt"), 0o755))

		src := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(src, "myapp"), []byte("hello"), 0o755))

		dest, err := resolveContainerRootPath(root, "opt")
		Expect(err).ToNot(HaveOccurred())
		absTo, err := normalizeDependencyImportDestination(src, dest)
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(dest))
	})

	It("contains a destination symlink that would escape the root mount", func() {
		base := GinkgoT().TempDir()
		root := filepath.Join(base, "root")
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(base, "outside"), 0o755))
		require.NoError(GinkgoT(), os.MkdirAll(root, 0o755))
		require.NoError(GinkgoT(), os.Symlink("../outside", filepath.Join(root, "bin")))

		absTo, err := resolveContainerRootPath(root, "bin")
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(filepath.Join(root, "outside")))
	})
})
