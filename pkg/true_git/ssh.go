package true_git

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Every git request over ssh pays a full handshake, which dominates the cost of
// small requests such as ls-remote. OpenSSH multiplexing reuses one connection
// per host, so the handshake is paid once.
const (
	sshControlPersist = "60s"

	// A unix socket path is limited to 104 bytes on macOS and 108 on Linux; ssh
	// expands %C to 40 hexadecimal characters.
	sshControlPathLimit = 100
)

var sshMultiplexingEnv []string

func setupSSHMultiplexing() []string {
	if runtime.GOOS == "windows" {
		return nil
	}

	if os.Getenv("GIT_SSH_COMMAND") != "" || os.Getenv("GIT_SSH") != "" {
		return nil
	}

	dir := filepath.Join(os.TempDir(), "werf-ssh")
	controlPath := filepath.Join(dir, "s-%C")
	if len(controlPath) > sshControlPathLimit {
		return nil
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil
	}

	return []string{fmt.Sprintf(`GIT_SSH_COMMAND=ssh -o ControlMaster=auto -o ControlPath="%s" -o ControlPersist=%s`, controlPath, sshControlPersist)}
}
