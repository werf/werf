package true_git

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ssh multiplexing", func() {
	gitSSHCommand := func(cmd GitCmd) string {
		var res string
		for _, entry := range cmd.Env {
			if value, ok := strings.CutPrefix(entry, "GIT_SSH_COMMAND="); ok {
				res = value
			}
		}
		return res
	}

	BeforeEach(func() {
		GinkgoT().Setenv("GIT_SSH_COMMAND", "")
		GinkgoT().Setenv("GIT_SSH", "")
	})

	It("reuses one connection per host", func(ctx SpecContext) {
		Expect(Init(ctx, Options{})).To(Succeed())

		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(command).To(ContainSubstring("ControlMaster=auto"))
		Expect(command).To(ContainSubstring(filepath.Join(os.TempDir(), "werf-ssh", "s-%C")))

		info, err := os.Stat(filepath.Join(os.TempDir(), "werf-ssh"))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o700)))
	})

	DescribeTable("keeps the ssh command chosen by the user", func(ctx SpecContext, variable string) {
		GinkgoT().Setenv(variable, "ssh -i /tmp/key")
		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).NotTo(ContainSubstring("ControlMaster"))
	},
		Entry("GIT_SSH_COMMAND", "GIT_SSH_COMMAND"),
		Entry("GIT_SSH", "GIT_SSH"),
	)

	It("lets the caller environment win", func(ctx SpecContext) {
		Expect(Init(ctx, Options{})).To(Succeed())

		cmd := NewGitCmd(ctx, &GitCmdOptions{Env: []string{"GIT_SSH_COMMAND=ssh -i /tmp/key"}}, "version")
		Expect(gitSSHCommand(cmd)).To(Equal("ssh -i /tmp/key"))
	})
})
