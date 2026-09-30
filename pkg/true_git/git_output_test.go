package true_git

import (
	"bytes"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.DescribeTable("Git ref output",
	func(ctx ginkgo.SpecContext, command string, trace bool) {
		ginkgo.GinkgoT().Setenv("WERF_DEBUG_TRUE_GIT", "0")
		if trace {
			ginkgo.GinkgoT().Setenv("WERF_DEBUG_TRUE_GIT", "1")
		}
		repo := filepath.Join(ginkgo.GinkgoT().TempDir(), "repo")
		utils.MkdirAll(repo)
		utils.RunSucceedCommand(ctx, repo, "git", "-c", "init.defaultBranch=main", "init")
		gitCommitSucceed(ctx, repo, "--allow-empty", "-m", "Initial commit")
		utils.RunSucceedCommand(ctx, repo, "git", "tag", "visible-tag")

		previousLiveOutput := liveGitOutput
		ginkgo.DeferCleanup(func() { liveGitOutput = previousLiveOutput })
		gomega.Expect(Init(ctx, Options{LiveGitOutput: true})).To(gomega.Succeed())
		var output bytes.Buffer
		ctxWithLogger := logboek.NewContext(ctx, logboek.NewLogger(&output, &output))
		expected := "refs/tags/visible-tag"
		args := []string{command, "--tags"}
		if command == "branch" {
			args = []string{command, "--list", "main"}
			expected = "main"
		}
		if command == "ls-remote" {
			args = append(args, repo)
		}
		cmd := NewGitCmd(ctxWithLogger, &GitCmdOptions{RepoDir: repo}, args...)
		gomega.Expect(cmd.Run(ctxWithLogger)).To(gomega.Succeed())
		gomega.Expect(cmd.OutBuf.String()).To(gomega.ContainSubstring(expected))
		gomega.Expect(cmd.OutErrBuf.String()).To(gomega.ContainSubstring(expected))
		if trace || command == "branch" {
			gomega.Expect(output.String()).To(gomega.ContainSubstring(expected))
		} else {
			gomega.Expect(output.String()).NotTo(gomega.ContainSubstring(expected))
		}
	},
	ginkgo.Entry("retain output of other Git commands", "branch", false),
	ginkgo.Entry("hide local refs from general debug output", "show-ref", false),
	ginkgo.Entry("hide remote refs from general debug output", "ls-remote", false),
	ginkgo.Entry("retain local refs in explicit Git tracing", "show-ref", true),
	ginkgo.Entry("retain remote refs in explicit Git tracing", "ls-remote", true),
)
