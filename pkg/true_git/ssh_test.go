package true_git

import (
	"os"
	"os/exec"
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
		DeferCleanup(CleanupSSHMultiplexing)
	})

	It("creates a private connection directory", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		Expect(Init(ctx, Options{})).To(Succeed())

		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(command).To(HavePrefix(`sh "`))
		Expect(controlDir(command)).To(HavePrefix(filepath.Join(os.TempDir(), "werf-ssh-")))

		info, err := os.Stat(controlDir(command))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o700)))
	})

	It("separates aliases with different keys while reusing connections across repositories", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		Expect(Init(ctx, Options{})).To(Succeed())
		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(command).NotTo(BeEmpty())

		config := filepath.Join(shortTempDir(), "config")
		longAlias := strings.Repeat("long", 60)
		Expect(os.WriteFile(config, []byte("Host first "+longAlias+"\n HostName github.com\n User git\n IdentityFile /first-key\nHost second\n HostName github.com\n User git\n IdentityFile /second-key\n"), 0o600)).To(Succeed())

		var sockets []string
		for _, connection := range []struct {
			alias, repo, key, port string
		}{
			{"first", "one", "/first-key", "22"},
			{"second", "one", "/second-key", "22"},
			{"first", "two", "/first-key", "22"},
			{longAlias, "one", "/first-key", "22"},
			{"first", "one", "/first-key", "2222"},
		} {
			cmd := exec.CommandContext(ctx, "sh", "-c", command+` "$@"`, "ssh", "-G", "-F", config, "-p", connection.port, connection.alias, "git-upload-pack '"+connection.repo+"'")
			output, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("identityfile " + connection.key + "\n"))
			Expect(string(output)).To(ContainSubstring("controlmaster auto\n"))
			Expect(string(output)).To(ContainSubstring("controlpersist 60\n"))
			var socket string
			for _, line := range strings.Split(string(output), "\n") {
				if value, ok := strings.CutPrefix(line, "controlpath "); ok {
					socket = value
				}
			}
			Expect(socket).NotTo(BeEmpty())
			Expect(len(socket) + len(".") + 16).To(BeNumerically("<=", sshControlPathLimit))
			sockets = append(sockets, socket)
		}
		Expect(sockets[0]).NotTo(Equal(sockets[1]))
		Expect(sockets[0]).To(Equal(sockets[2]))
		Expect(sockets[0]).NotTo(Equal(sockets[3]))
		Expect(sockets[0]).NotTo(Equal(sockets[4]))
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

		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).To(BeEmpty())
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

	It("reclaims its control directory on cleanup", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		Expect(Init(ctx, Options{})).To(Succeed())

		dir := controlDir(gitSSHCommand(NewGitCmd(ctx, nil, "version")))
		Expect(dir).To(BeADirectory())

		CleanupSSHMultiplexing()
		Expect(dir).NotTo(BeADirectory())
		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).To(BeEmpty())
	})

	It("stays off when the ssh binary does not understand multiplexing", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())

		binDir := shortTempDir()
		fakeSSH := filepath.Join(binDir, "ssh")
		Expect(os.WriteFile(fakeSSH, []byte("#!/bin/sh\nexit 1\n"), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).To(BeEmpty())
	})

	DescribeTable("uses plain ssh when hashing fails", func(ctx SpecContext, hashScript string) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		binDir := shortTempDir()
		Expect(os.WriteFile(filepath.Join(binDir, "sha256sum"), []byte("#!/bin/sh\n"+hashScript), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		Expect(Init(ctx, Options{})).To(Succeed())

		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(command).NotTo(BeEmpty())
		cmd := exec.CommandContext(ctx, "sh", "-c", command+` "$@"`, "ssh", "-G", "-F", os.DevNull, "werf-probe", "true")
		output, err := cmd.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("hostname werf-probe\n"))
		Expect(string(output)).To(ContainSubstring("controlmaster false\n"))
		Expect(string(output)).NotTo(ContainSubstring("controlpath "))
	},
		Entry("failed command", "printf '%064d\\n' 0\nexit 1\n"),
		Entry("empty digest", "exit 0\n"),
		Entry("non-hex digest", "printf '%064d\\n' 0 | tr 0 z\n"),
		Entry("short digest", "printf 'abcd\\n'\n"),
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
	_, path, _ := strings.Cut(sshCommand, `sh "`)
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
