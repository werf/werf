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

		absTo, err := normalizeDependencyImportDestination(root, src, filepath.Join(root, "bin"))
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

		absTo, err := normalizeDependencyImportDestination(root, srcFile, filepath.Join(root, "bin"))
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(filepath.Join(root, "usr", "bin", "myapp")))
	})

	It("anchors an absolute symlink destination target under the root mount", func() {
		root := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(root, "usr", "bin"), 0o755))
		require.NoError(GinkgoT(), os.Symlink("/usr/bin", filepath.Join(root, "bin")))

		src := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(src, "myapp"), []byte("hello"), 0o755))

		absTo, err := normalizeDependencyImportDestination(root, src, filepath.Join(root, "bin"))
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(filepath.Join(root, "usr", "bin")))
	})

	It("leaves a non-symlink directory destination unchanged", func() {
		root := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(root, "opt"), 0o755))

		src := GinkgoT().TempDir()
		require.NoError(GinkgoT(), os.WriteFile(filepath.Join(src, "myapp"), []byte("hello"), 0o755))

		dest := filepath.Join(root, "opt")
		absTo, err := normalizeDependencyImportDestination(root, src, dest)
		Expect(err).ToNot(HaveOccurred())
		Expect(absTo).To(Equal(dest))
	})

	It("rejects a destination symlink that escapes the root mount", func() {
		base := GinkgoT().TempDir()
		root := filepath.Join(base, "root")
		require.NoError(GinkgoT(), os.MkdirAll(filepath.Join(base, "outside"), 0o755))
		require.NoError(GinkgoT(), os.MkdirAll(root, 0o755))
		require.NoError(GinkgoT(), os.Symlink("../outside", filepath.Join(root, "bin")))

		_, err := resolveDestSymlinkUnderRoot(root, filepath.Join(root, "bin"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("escapes root mount"))
	})
})
