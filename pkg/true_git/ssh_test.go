package true_git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
)

var _ = Describe("Ssh multiplexing", func() {
	gitSSHCommand := func(cmd GitCmd) string {
		var res string
		for _, entry := range cmd.Env {
			if value, ok := strings.CutPrefix(entry, "GIT_SSH="); ok {
				res = value
			}
			if value, ok := strings.CutPrefix(entry, "GIT_SSH_COMMAND="); ok {
				res = value
			}
		}
		return res
	}

	BeforeEach(func() {
		GinkgoT().Setenv("GIT_SSH_COMMAND", "")
		GinkgoT().Setenv("GIT_SSH", "")
		Expect(os.Unsetenv("GIT_SSH_COMMAND")).To(Succeed())
		Expect(os.Unsetenv("GIT_SSH")).To(Succeed())
		GinkgoT().Setenv("GIT_CONFIG_GLOBAL", filepath.Join(shortTempDir(), "gitconfig"))
		GinkgoT().Setenv("GIT_CONFIG_SYSTEM", filepath.Join(shortTempDir(), "gitconfig"))
		DeferCleanup(CleanupSSHMultiplexing)
	})

	It("creates a private connection directory", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		Expect(Init(ctx, Options{})).To(Succeed())

		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(filepath.Base(command)).To(Equal("ssh"))
		wrapperInfo, err := os.Stat(command)
		Expect(err).NotTo(HaveOccurred())
		Expect(wrapperInfo.Mode().Perm()).To(Equal(os.FileMode(0o700)))
		Expect(controlDir(command)).To(HavePrefix(filepath.Join(os.TempDir(), "werf-ssh-")))

		info, err := os.Stat(controlDir(command))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o700)))
	})

	DescribeTable("separates aliases with different keys while reusing connections across repositories", func(ctx SpecContext, shell string) {
		if _, err := exec.LookPath(shell); err != nil {
			Skip("shell not installed: " + shell)
		}
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
			args := []string{command, "-G", "-F", config, "-p", connection.port, connection.alias, "git-upload-pack '" + connection.repo + "'"}
			if shell == "busybox" {
				args = append([]string{"sh"}, args...)
			}
			cmd := exec.CommandContext(ctx, shell, args...)
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
			Expect(socket).To(HavePrefix(filepath.Join(sshControlDir, "s-")))
			Expect(len(socket) + len(".") + 16).To(BeNumerically("<=", sshControlPathLimit))
			sockets = append(sockets, socket)
		}
		Expect(sockets[0]).NotTo(Equal(sockets[1]))
		Expect(sockets[0]).To(Equal(sockets[2]))
		Expect(sockets[0]).NotTo(Equal(sockets[3]))
		Expect(sockets[0]).NotTo(Equal(sockets[4]))
	},
		Entry("system sh", "/bin/sh"),
		Entry("bash", "bash"),
		Entry("dash", "dash"),
		Entry("BusyBox ash", "busybox"),
	)

	DescribeTable("lets git recognize ssh while respecting the user's variant", func(ctx SpecContext, envVariant, configVariant string, sendProtocol bool) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		GinkgoT().Setenv("GIT_SSH_VARIANT", "")
		Expect(os.Unsetenv("GIT_SSH_VARIANT")).To(Succeed())
		if envVariant != "" {
			GinkgoT().Setenv("GIT_SSH_VARIANT", envVariant)
		}
		if configVariant != "" {
			Expect(os.WriteFile(os.Getenv("GIT_CONFIG_GLOBAL"), []byte("[ssh]\nvariant = "+configVariant+"\n"), 0o600)).To(Succeed())
		}
		Expect(Init(ctx, Options{})).To(Succeed())

		binDir := shortTempDir()
		calls := filepath.Join(binDir, "calls")
		Expect(os.WriteFile(filepath.Join(binDir, "ssh"), []byte("#!/bin/sh\nprintf 'call\\n' >> \"$WERF_TEST_SSH_CALLS\"\nprintf '%s\\n' \"$@\" >> \"$WERF_TEST_SSH_CALLS\"\nexit 0\n"), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("WERF_TEST_SSH_CALLS", calls)

		cmd := NewGitCmd(ctx, nil, "-c", "protocol.version=2", "ls-remote", "git@werf-probe:/repo")
		Expect(cmd.Run(ctx)).NotTo(Succeed())
		output, err := os.ReadFile(calls)
		Expect(err).NotTo(HaveOccurred())
		args := strings.Split(strings.TrimSpace(string(output)), "\n")
		Expect(args).To(ContainElement("git-upload-pack '/repo'"))
		Expect(strings.Count(string(output), "call\n")).To(Equal(1))
		Expect(args).NotTo(ContainElement("-G"))
		Expect(strings.Contains(string(output), "\nSendEnv=GIT_PROTOCOL\n")).To(Equal(sendProtocol))
	},
		Entry("no extra probe by default", "", "", true),
		Entry("environment variant", "simple", "", false),
		Entry("configured variant", "", "simple", false),
		Entry("environment takes precedence", "ssh", "simple", true),
		Entry("explicit autodetection", "auto", "", true),
	)

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

	DescribeTable("keeps the target repository SSH transport", func(ctx SpecContext, conditionalInclude bool) {
		root := shortTempDir()
		Expect(Init(ctx, Options{})).To(Succeed())
		Expect(sshMultiplexingEnv).NotTo(BeEmpty())

		repo := filepath.Join(root, "project")
		gitInitRepo(ctx, repo)
		custom := filepath.Join(root, "custom-ssh")
		marker := filepath.Join(root, "custom-called")
		Expect(os.WriteFile(custom, []byte("#!/bin/sh\ntouch '"+marker+"'\nexit 1\n"), 0o700)).To(Succeed())
		if conditionalInclude {
			included := filepath.Join(root, "included-config")
			Expect(os.WriteFile(included, []byte("[core]\nsshCommand = "+custom+"\n"), 0o600)).To(Succeed())
			Expect(os.WriteFile(os.Getenv("GIT_CONFIG_GLOBAL"), []byte("[includeIf \"gitdir:**/project/\"]\npath = "+included+"\n"), 0o600)).To(Succeed())
		} else {
			gitSucceed(ctx, repo, "config", "core.sshCommand", custom)
		}

		bin := filepath.Join(root, "bin")
		Expect(os.Mkdir(bin, 0o700)).To(Succeed())
		fallbackMarker := filepath.Join(root, "fallback-called")
		Expect(os.WriteFile(filepath.Join(bin, "ssh"), []byte("#!/bin/sh\ntouch '"+fallbackMarker+"'\nexit 1\n"), 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

		cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repo}, "ls-remote", "git@werf-probe:/repo")
		Expect(cmd.Run(ctx)).NotTo(Succeed())
		Expect(marker).To(BeAnExistingFile())
		Expect(fallbackMarker).NotTo(BeAnExistingFile())
	},
		Entry("local config", false),
		Entry("conditional include", true),
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
		Entry("the directory is relative to the process working directory", func() string {
			dir, err := os.MkdirTemp(".", "t")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(os.RemoveAll(dir)).To(Succeed()) })
			return dir
		}),
		Entry("the directory cannot be created", func() string {
			path := filepath.Join(shortTempDir(), "file")
			Expect(os.WriteFile(path, nil, 0o600)).To(Succeed())
			return path
		}),
	)

	DescribeTable("avoids shell syntax and SSH tokens in the temporary path", func(ctx SpecContext, character string) {
		base := filepath.Join(shortTempDir(), character)
		Expect(os.Mkdir(base, 0o700)).To(Succeed())
		GinkgoT().Setenv("TMPDIR", base)
		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(controlDir(gitSSHCommand(NewGitCmd(ctx, nil, "version")))).To(HavePrefix(filepath.Join("/tmp", "werf-ssh-")))
	},
		Entry("single quote", "'"),
		Entry("double quote", "\""),
		Entry("dollar", "$"),
		Entry("backtick", "`"),
		Entry("backslash", "\\"),
		Entry("percent", "%"),
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

		CleanupSSHMultiplexing(ctx)
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

	DescribeTable("uses an available SHA-256 implementation", func(specCtx SpecContext, available, expected []string) {
		var log bytes.Buffer
		logger := logboek.NewLogger(&log, &log)
		logger.SetAcceptedLevel(level.Default)
		ctx := logboek.NewContext(specCtx, logger)

		GinkgoT().Setenv("TMPDIR", shortTempDir())
		binDir := shortTempDir()
		for _, name := range []string{"git", "ssh", "sh"} {
			path, err := exec.LookPath(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Symlink(path, filepath.Join(binDir, name))).To(Succeed())
		}
		calls := filepath.Join(binDir, "hash-args")
		GinkgoT().Setenv("WERF_TEST_SSH_HASH_ARGS", calls)
		for _, name := range available {
			script := "#!/bin/sh\nprintf '%s\\n' '" + name + "' \"$@\" > \"$WERF_TEST_SSH_HASH_ARGS\"\nprintf '%064d *stdin\\n' 0\n"
			Expect(os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755)).To(Succeed())
		}
		GinkgoT().Setenv("PATH", binDir)
		Expect(Init(ctx, Options{})).To(Succeed())
		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		if len(expected) == 0 {
			Expect(command).To(BeEmpty())
			Expect(log.String()).To(ContainSubstring("sha256sum, openssl or shasum"))
			return
		}
		Expect(command).NotTo(BeEmpty())
		Expect(log.String()).NotTo(ContainSubstring("sha256sum, openssl or shasum"))
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command+` "$@"`, "ssh", "-G", "-F", os.DevNull, "werf-probe", "true")
		output, err := cmd.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("controlpath " + filepath.Join(sshControlDir, "s-"+strings.Repeat("0", 40)) + "\n"))
		args, err := os.ReadFile(calls)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Fields(string(args))).To(Equal(expected))
	},
		Entry("prefers sha256sum", []string{"sha256sum", "openssl", "shasum"}, []string{"sha256sum"}),
		Entry("falls back to openssl", []string{"openssl", "shasum"}, []string{"openssl", "dgst", "-sha256", "-r"}),
		Entry("falls back to shasum", []string{"shasum"}, []string{"shasum", "-a", "256"}),
		Entry("stays off without a hash utility", []string{}, []string{}),
	)

	It("stays off when the installed hash utility fails its initial probe", func(ctx SpecContext) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		binDir := shortTempDir()
		Expect(os.WriteFile(filepath.Join(binDir, "sha256sum"), []byte("#!/bin/sh\nexit 1\n"), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		Expect(Init(ctx, Options{})).To(Succeed())

		Expect(gitSSHCommand(NewGitCmd(ctx, nil, "version"))).To(BeEmpty())
	})

	DescribeTable("disables multiplexing when hashing fails after initialization", func(ctx SpecContext, hashScript string) {
		GinkgoT().Setenv("TMPDIR", shortTempDir())
		binDir := shortTempDir()
		hashPath := filepath.Join(binDir, "sha256sum")
		Expect(os.WriteFile(hashPath, []byte("#!/bin/sh\nprintf '%064d\\n' 0\n"), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		Expect(Init(ctx, Options{})).To(Succeed())

		command := gitSSHCommand(NewGitCmd(ctx, nil, "version"))
		Expect(command).NotTo(BeEmpty())
		Expect(os.WriteFile(hashPath, []byte("#!/bin/sh\n"+hashScript), 0o755)).To(Succeed())
		cmd := exec.CommandContext(ctx, "sh", "-c", command+` "$@"`, "ssh", "-G", "-F", os.DevNull, "werf-probe", "true")
		output, err := cmd.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("hostname werf-probe\n"))
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
	return filepath.Dir(sshCommand)
}

// The control path limit rejects a long directory before anything else, and
// the directory a test framework hands out is already long.
func shortTempDir() string {
	dir, err := os.MkdirTemp("/tmp", "t")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(os.RemoveAll(dir)).To(Succeed()) })

	return dir
}
