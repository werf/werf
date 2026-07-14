package true_git

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func runGitAI(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return string(out)
}

func initGitRepoAI(t *testing.T, dir string) {
	t.Helper()
	runGitAI(t, dir, "init")
	runGitAI(t, dir, "config", "user.email", "test@werf.io")
	runGitAI(t, dir, "config", "user.name", "test")
	runGitAI(t, dir, "commit", "--allow-empty", "-m", "init")
}
