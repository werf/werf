package tmp_manager

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tmp manager", func() {
	Describe("gc", func() {
		DescribeTable("list and filter paths",
			func(setup func(linkedDir, targetDir string) (string, []string, []string)) {
				linkedDir := GinkgoT().TempDir()
				targetDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("%s-target", filepath.Base(linkedDir)))
				Expect(err).To(Succeed())
				defer os.RemoveAll(targetDir)

				linkedDir, expectedFiles, expectedSymlinks := setup(linkedDir, targetDir)
				actualFiles, actualSymlinks, err := listDirAndFollowSymlinks(linkedDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualFiles).To(Equal(expectedFiles))
				Expect(actualSymlinks).To(Equal(expectedSymlinks))
			},
			Entry("linked directory not exist",
				func(linkedDir, targetDir string) (string, []string, []string) {
					return "some-not-existing-linked-dir", nil, nil
				}),
			Entry("linked directory contains no files",
				func(linkedDir, targetDir string) (string, []string, []string) {
					// Nothing to do since directory is empty
					return linkedDir, []string{}, []string{}
				}),
			Entry("linked directory contains file created any time ago hours ago",
				func(linkedDir, targetDir string) (string, []string, []string) {
					file1 := filepath.Join(linkedDir, "file.txt")
					Expect(os.WriteFile(file1, []byte("file"), 0o644)).To(Succeed())

					return linkedDir, []string{filepath.Join(linkedDir, "file.txt")}, []string{}
				}),
			Entry("linked directory contains one symlink any time ago and target dir contains one file",
				func(linkedDir, targetDir string) (string, []string, []string) {
					targetFile := filepath.Join(targetDir, "target.txt")
					Expect(os.WriteFile(targetFile, []byte("target"), 0o644)).To(Succeed())

					symlink := filepath.Join(linkedDir, "symlink")
					Expect(os.Symlink(targetFile, symlink)).To(Succeed())

					return linkedDir, []string{filepath.Join(targetDir, "target.txt")}, []string{filepath.Join(linkedDir, "symlink")}
				},
			),
			Entry("linked directory contains one (broken) symlink created any time ago and target file does not exist",
				func(linkedDir, targetDir string) (string, []string, []string) {
					targetFile := filepath.Join(targetDir, "target.txt")
					Expect(os.WriteFile(targetFile, []byte("target"), 0o644)).To(Succeed())

					symlink := filepath.Join(linkedDir, "symlink")
					Expect(os.Symlink(targetFile, symlink)).To(Succeed())

					Expect(os.RemoveAll(targetFile)).To(Succeed())

					return linkedDir, []string{}, []string{filepath.Join(linkedDir, "symlink")}
				},
			),
		)
	})
})
