package true_git

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func runGitAI(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return string(out)
}

// isolateGitConfigAI detaches the test from the developer's global/system git config, so an ambient
// setting (notably protocol.file.allow=always) cannot mask a missing production option.
func isolateGitConfigAI(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
}

func initGitRepoAI(t *testing.T, dir string) {
	t.Helper()
	runGitAI(t, dir, "init")
	runGitAI(t, dir, "config", "user.email", "test@werf.io")
	runGitAI(t, dir, "config", "user.name", "test")
	runGitAI(t, dir, "commit", "--allow-empty", "-m", "init")
}
