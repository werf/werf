package tmp_manager

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
)

var _ = Describe("tmp manager", func() {
	Describe("gc", func() {
		DescribeTable("list and filter paths",
			func(setup func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string)) {
				linkedDir := GinkgoT().TempDir()
				targetDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("%s-target", filepath.Base(linkedDir)))
				Expect(err).To(Succeed())
				defer os.RemoveAll(targetDir)
				// There is no way to change file modification time while testing.
				// A workaround is using stubs.
				stubs := gostub.New()
				defer stubs.Reset()

				linkedDir, expectedFiles, expectedSymlinks := setup(linkedDir, targetDir, stubs)
				actualFiles, actualSymlinks, err := listDirAndFollowSymlinks(linkedDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualFiles).To(Equal(expectedFiles))
				Expect(actualSymlinks).To(Equal(expectedSymlinks))
			},
			Entry("linked directory not exist",
				func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string) {
					return "some-not-existing-linked-dir", nil, nil
				}),
			Entry("linked directory contains no files",
				func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string) {
					// Nothing to do since directory is empty
					return linkedDir, []string{}, []string{}
				}),
			Entry("linked directory contains file created 3 hours ago",
				func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string) {
					file1 := filepath.Join(linkedDir, "file.txt")
					Expect(os.WriteFile(file1, []byte("file"), 0o644)).To(Succeed())
					stubs.StubFunc(&timeSince, time.Hour*3)

					return linkedDir, []string{filepath.Join(linkedDir, "file.txt")}, []string{}
				}),
			Entry("linked directory contains file created 1 hours ago",
				func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string) {
					file1 := filepath.Join(linkedDir, "file.txt")
					Expect(os.WriteFile(file1, []byte("file"), 0o644)).To(Succeed())
					stubs.StubFunc(&timeSince, time.Hour*1)

					return linkedDir, []string{}, []string{}
				}),
			Entry("linked directory contains one symlink created 3 hours ago and target dir contains one file",
				func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string) {
					targetFile := filepath.Join(targetDir, "target.txt")
					Expect(os.WriteFile(targetFile, []byte("target"), 0o644)).To(Succeed())

					symlink := filepath.Join(linkedDir, "symlink")
					Expect(os.Symlink(targetFile, symlink)).To(Succeed())
					stubs.StubFunc(&timeSince, time.Hour*3)

					return linkedDir, []string{filepath.Join(targetDir, "target.txt")}, []string{filepath.Join(linkedDir, "symlink")}
				},
			),
			Entry("linked directory contains one (broken) symlink created 3 hours and target file does not exist",
				func(linkedDir, targetDir string, stubs *gostub.Stubs) (string, []string, []string) {
					targetFile := filepath.Join(targetDir, "target.txt")
					Expect(os.WriteFile(targetFile, []byte("target"), 0o644)).To(Succeed())

					symlink := filepath.Join(linkedDir, "symlink")
					Expect(os.Symlink(targetFile, symlink)).To(Succeed())
					stubs.StubFunc(&timeSince, time.Hour*3)

					Expect(os.RemoveAll(targetFile)).To(Succeed())

					return linkedDir, []string{}, []string{filepath.Join(linkedDir, "symlink")}
				},
			),
		)
	})
})
