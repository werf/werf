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
		GinkgoT().Setenv("GIT_CONFIG_GLOBAL", filepath.Join(shortTempDir(), "gitconfig"))
		GinkgoT().Setenv("GIT_CONFIG_SYSTEM", filepath.Join(shortTempDir(), "gitconfig"))
	})

	It("reuses one connection per host", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		Expect(Init(ctx, Options{})).To(Succeed())

		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(command).To(ContainSubstring("ControlMaster=auto"))
		Expect(controlDir(command)).To(HavePrefix(filepath.Join(os.TempDir(), "werf-ssh-")))

		info, err := os.Stat(controlDir(command))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o700)))
	})

	It("gives every werf process its own socket", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())

		Expect(Init(ctx, Options{})).To(Succeed())
		first := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(Init(ctx, Options{})).To(Succeed())
		second := gitSSHCommand(NewGitCmd(ctx, nil, "version"))

		Expect(first).NotTo(Equal(second))
	})

	DescribeTable("keeps the ssh command chosen by the user", func(ctx SpecContext, setup func()) {
		setup()
		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).NotTo(ContainSubstring("ControlMaster"))
	},
		Entry("GIT_SSH_COMMAND", func() { GinkgoT().Setenv("GIT_SSH_COMMAND", "ssh -i /tmp/key") }),
		Entry("GIT_SSH", func() { GinkgoT().Setenv("GIT_SSH", "ssh -i /tmp/key") }),
		Entry("core.sshCommand", func() {
			path := filepath.Join(shortTempDir(), "gitconfig")
			Expect(os.WriteFile(path, []byte("[core]\n\tsshCommand = ssh -i /tmp/key\n"), 0o600)).To(Succeed())
			GinkgoT().Setenv("GIT_CONFIG_GLOBAL", path)
		}),
	)

	It("lets the caller environment win", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		Expect(Init(ctx, Options{})).To(Succeed())

		cmd := NewGitCmd(ctx, &GitCmdOptions{Env: []string{"GIT_SSH_COMMAND=ssh -i /tmp/key"}}, "version")
		Expect(gitSSHCommand(cmd)).To(Equal("ssh -i /tmp/key"))
	})

	DescribeTable("falls back to /tmp when the temporary directory cannot hold the socket", func(ctx SpecContext, tmpDir func() string) {
		GinkgoT().Setenv("TMPDIR", tmpDir())
		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(controlDir(gitSSHCommand(NewGitCmd(ctx, nil, "version")))).To(HavePrefix(filepath.Join("/tmp", "werf-ssh-")))
	},
		Entry("the socket ssh binds would be too long", func() string {
			// Long enough that the path ssh binds under does not fit, short
			// enough that the control path itself does.
			dir := filepath.Join("/tmp", strings.Repeat("d", 30))
			Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
			DeferCleanup(func() { Expect(os.RemoveAll(dir)).To(Succeed()) })
			return dir
		}),
		Entry("the directory cannot be created", func() string {
			path := filepath.Join(shortTempDir(), "file")
			Expect(os.WriteFile(path, nil, 0o600)).To(Succeed())
			return path
		}),
	)

	DescribeTable("refuses a directory that cannot hold the socket ssh creates", func(setup func(dir string), expected bool) {
		dir := shortTempDir()
		setup(dir)

		Expect(canHoldControlSocket(dir)).To(Equal(expected))
	},
		Entry("an ordinary directory", func(string) {}, true),
		Entry("the socket cannot be linked into place", func(dir string) {
			Expect(os.MkdirAll(filepath.Join(dir, "probe"), 0o700)).To(Succeed())
		}, false),
		Entry("the socket cannot be bound", func(dir string) {
			Expect(os.MkdirAll(filepath.Join(dir, "probe.tmp"), 0o700)).To(Succeed())
		}, false),
	)

	It("gives up when no directory can hold the socket", func(ctx SpecContext) {
		path := filepath.Join(shortTempDir(), "file")
		Expect(os.WriteFile(path, nil, 0o600)).To(Succeed())
		GinkgoT().Setenv("TMPDIR", path)
		sshFallbackDir = path
		DeferCleanup(func() { sshFallbackDir = "/tmp" })

		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).To(BeEmpty())
	})
})

func controlDir(sshCommand string) string {
	_, path, _ := strings.Cut(sshCommand, `ControlPath="`)
	path, _, _ = strings.Cut(path, `"`)

	return filepath.Dir(path)
}

// The control path limit rejects a long directory before anything else, and
// the directory a test framework hands out is already long.
func shortTempDir() string {
	dir, err := os.MkdirTemp("/tmp", "t")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(os.RemoveAll(dir)).To(Succeed()) })

	return dir
}
