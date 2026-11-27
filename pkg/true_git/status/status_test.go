package status

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status", func() {
	DescribeTable("returns file names exactly as they are on disk",
		func(ctx context.Context, name string) {
			dir := GinkgoT().TempDir()
			git := func(args ...string) {
				cmd := exec.CommandContext(ctx, "git", append([]string{"-c", "user.name=werf", "-c", "user.email=werf@example.com", "-c", "commit.gpgsign=false"}, args...)...)
				cmd.Dir = dir
				output, err := cmd.CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))
			}
			write := func(path, data string) {
				Expect(os.WriteFile(filepath.Join(dir, path), []byte(data), 0o600)).To(Succeed())
			}

			tracked := "tracked " + name
			untracked := "untracked " + name

			git("init")
			write(tracked, "one")
			git("add", ".")
			git("commit", "-m", "init")
			write(tracked, "two")
			write(untracked, "new")

			result, err := Status(ctx, dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Worktree.PathList()).To(Equal([]string{tracked}))
			Expect(result.UntrackedPathList).To(Equal([]string{untracked}))

			git("add", tracked)
			result, err = Status(ctx, dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Index.PathList()).To(Equal([]string{tracked}))
		},
		Entry("plain name", "file.yaml"),
		Entry("double quote", `a"b.yaml`),
		Entry("backslash", `a\b.yaml`),
		Entry("tab", "a\tb.yaml"),
		Entry("newline", "a\nb.yaml"),
		Entry("consecutive spaces", "a  b.yaml"),
		Entry("non-ASCII", "ñ.yaml"),
	)
})
