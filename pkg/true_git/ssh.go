package true_git

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Every git request over ssh pays a full handshake, which dominates the cost of
// small requests such as ls-remote. OpenSSH multiplexing reuses one connection
// per SSH alias and connection options, so the handshake is paid once.
const (
	sshControlPersist = "60s"

	// A unix socket path is limited to 104 bytes on macOS and 108 on Linux, and
	// what has to fit is the path ssh binds before it links the socket into
	// place: the control path with a 40-character hash, plus a dot and
	// 16 random characters.
	sshControlPathLimit = 100
	sshControlSuffixLen = 40 + len(".") + 16
)

var (
	sshMultiplexingEnv []string
	sshControlDir      string
	sshFallbackDir     = "/tmp"
)

func setupSSHMultiplexing(ctx context.Context) []string {
	if runtime.GOOS == "windows" {
		return nil
	}

	// git gives GIT_SSH_COMMAND precedence over core.sshCommand, so setting it
	// would silently discard the ssh command, identity or proxy a user
	// configured in the environment, the global or system config, or the
	// repository the process runs in. A repository named with --dir is read
	// like any other repository werf fetches into: its local config is not
	// consulted here.
	if os.Getenv("GIT_SSH_COMMAND") != "" || os.Getenv("GIT_SSH") != "" || configuredSSHCommand(ctx) != "" {
		return nil
	}

	hashCommand := "sha256sum"
	if _, err := exec.LookPath(hashCommand); err != nil {
		if _, err := exec.LookPath("shasum"); err != nil {
			return nil
		}
		hashCommand = "shasum -a 256"
	}

	for _, base := range []string{os.TempDir(), sshFallbackDir} {
		// The control path is interpolated into a command git hands to the
		// shell, so a base directory with shell-active characters cannot be
		// carried safely.
		if strings.ContainsAny(base, "\"$`\\") {
			continue
		}

		// The directory is private to this werf process: a predictable one is
		// a socket another user can pre-create and answer on, and a shared one
		// hands the ssh connection of one build, authenticated with its own
		// keys, to the next build that asks for the same host.
		dir, err := os.MkdirTemp(base, "werf-ssh-")
		if err != nil {
			continue
		}

		controlPath := filepath.Join(dir, "s-")
		if len(controlPath)+sshControlSuffixLen > sshControlPathLimit || !canHoldControlSocket(dir) {
			os.RemoveAll(dir)
			continue
		}

		// Git appends the remote command last. Exclude it so repositories using
		// the same alias share a connection, but hash the original alias and
		// options: OpenSSH's %C loses aliases with distinct authentication keys.
		script := fmt.Sprintf(`hash=$(
	while [ "$#" -gt 1 ]; do
		printf '%%s\000' "$1"
		shift
	done | %s
) || exec ssh "$@"
hash=${hash%%%% *}
case "$hash" in
	''|*[!0-9a-f]*) exec ssh "$@" ;;
esac
[ "${#hash}" = 64 ] || exec ssh "$@"
hash=$(printf '%%.40s' "$hash")
exec ssh -o ControlMaster=auto -o ControlPath="%s$hash" -o ControlPersist=%s "$@"
`, hashCommand, controlPath, sshControlPersist)
		wrapper := filepath.Join(dir, "ssh")
		if err := os.WriteFile(wrapper, []byte(script), 0o600); err != nil {
			os.RemoveAll(dir)
			continue
		}

		if !sshSupportsMultiplexing(ctx, wrapper) {
			os.RemoveAll(dir)
			return nil
		}

		sshControlDir = dir
		return []string{fmt.Sprintf(`GIT_SSH_COMMAND=sh "%s"`, wrapper)}
	}

	return nil
}

// CleanupSSHMultiplexing removes the control directory of this process. The
// multiplexing master survives it for up to ControlPersist after its last
// client and exits on its own; only the directory needs reclaiming. A killed
// process leaves its directory behind: any liveness heuristic cheap enough to
// run on start risks deleting the socket of a build that is still running.
func CleanupSSHMultiplexing() {
	if sshControlDir == "" {
		return
	}

	os.RemoveAll(sshControlDir)
	sshControlDir = ""
	sshMultiplexingEnv = nil
}

// A non-OpenSSH ssh binary fails outright on unknown -o options, taking every
// git command down with it, so multiplexing is only enabled once the binary
// accepts them. -G resolves the configuration without connecting, and
// -F /dev/null keeps the probe independent of user and system configuration,
// where directives like CanonicalizeHostname could make it resolve names.
func sshSupportsMultiplexing(ctx context.Context, wrapper string) bool {
	cmd := exec.CommandContext(ctx, "sh", wrapper, "-G", "-F", "/dev/null", "werf-probe", "true")
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run() == nil
}

func configuredSSHCommand(ctx context.Context) string {
	cmd := NewGitCmd(ctx, nil, "config", "--get", "core.sshCommand")
	if err := cmd.Run(ctx); err != nil {
		return ""
	}

	return strings.TrimSpace(cmd.OutBuf.String())
}

// A directory can hold files but not a multiplexing socket — a bind mount of a
// host directory is the common case — and ssh reacts to that by failing the git
// command rather than by running without multiplexing, so the way ssh creates
// its socket is repeated here first: bind under a temporary name, then link.
func canHoldControlSocket(dir string) bool {
	bound := filepath.Join(dir, "probe.tmp")
	listener, err := net.Listen("unix", bound)
	if err != nil {
		return false
	}
	defer listener.Close()

	linked := filepath.Join(dir, "probe")
	if err := os.Link(bound, linked); err != nil {
		return false
	}

	return os.Remove(linked) == nil
}
