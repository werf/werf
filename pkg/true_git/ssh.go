package true_git

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Every git request over ssh pays a full handshake, which dominates the cost of
// small requests such as ls-remote. OpenSSH multiplexing reuses one connection
// per host, so the handshake is paid once.
const (
	sshControlPersist = "60s"

	// A unix socket path is limited to 104 bytes on macOS and 108 on Linux, and
	// what has to fit is the path ssh binds before it links the socket into
	// place: the control path with %C expanded to 40 characters, plus a dot and
	// 16 random characters.
	sshControlPathLimit = 100
	sshControlSuffixLen = 40 + len(".") + 16
)

var (
	sshMultiplexingEnv []string
	sshFallbackDir     = "/tmp"
)

func setupSSHMultiplexing(ctx context.Context) []string {
	if runtime.GOOS == "windows" {
		return nil
	}

	// git gives GIT_SSH_COMMAND precedence over core.sshCommand, so setting it
	// would silently discard the ssh command, identity or proxy a user
	// configured for this repository.
	if os.Getenv("GIT_SSH_COMMAND") != "" || os.Getenv("GIT_SSH") != "" || configuredSSHCommand(ctx) != "" {
		return nil
	}

	for _, base := range []string{os.TempDir(), sshFallbackDir} {
		// The directory is private to this werf process: a predictable one is
		// a socket another user can pre-create and answer on, and a shared one
		// hands the ssh connection of one build, authenticated with its own
		// keys, to the next build that asks for the same host.
		dir, err := os.MkdirTemp(base, "werf-ssh-")
		if err != nil {
			continue
		}

		controlPath := filepath.Join(dir, "s-%C")
		if len(controlPath)-len("%C")+sshControlSuffixLen > sshControlPathLimit || !canHoldControlSocket(dir) {
			os.RemoveAll(dir)
			continue
		}

		return []string{fmt.Sprintf(`GIT_SSH_COMMAND=ssh -o ControlMaster=auto -o ControlPath="%s" -o ControlPersist=%s`, controlPath, sshControlPersist)}
	}

	return nil
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
