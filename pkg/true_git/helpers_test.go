package true_git

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

// gitCommitSucceed pins identity and disables signing so commits do not depend on ambient
// global git config: a developer's commit.gpgsign needs a working gpg-agent, which
// intermittently fails under Ginkgo's parallel procs, and specs that isolate themselves via
// GIT_CONFIG_GLOBAL lose the only user.name/user.email the CI runner has.
func gitCommitSucceed(ctx context.Context, workTreeDir string, args ...string) {
	utils.RunSucceedCommand(ctx, workTreeDir, "git", append([]string{
		"-c", "commit.gpgsign=false",
		"-c", "user.email=werf@werf.io",
		"-c", "user.name=werf",
		"commit",
	}, args...)...)
}

func gitSucceed(ctx context.Context, dir string, args ...string) string {
	return utils.SucceedCommandOutputString(ctx, dir, "git", append([]string{
		"-c", "commit.gpgsign=false",
		"-c", "user.email=werf@werf.io",
		"-c", "user.name=werf",
	}, args...)...)
}

func gitSucceedTrimmed(ctx context.Context, dir string, args ...string) string {
	return strings.TrimSpace(gitSucceed(ctx, dir, args...))
}

func gitSucceedWithStdin(ctx context.Context, dir, stdin string, args ...string) string {
	out, err := utils.RunCommandWithOptions(ctx, dir, "git", args, utils.RunCommandOptions{
		ToStdin:       stdin,
		ShouldSucceed: true,
		NoStderr:      true,
	})
	Expect(err).ToNot(HaveOccurred())
	return strings.TrimSpace(string(out))
}

func gitInitRepo(ctx context.Context, dir string) {
	utils.MkdirAll(dir)
	gitSucceed(ctx, dir, "-c", "init.defaultBranch=main", "init")
	gitSucceed(ctx, dir, "commit", "--allow-empty", "-m", "init")
}

// setEnvForSpec sets an environment variable for the duration of the current spec. Specs of one
// Ginkgo process run serially, so a process-wide variable restored on teardown stays contained.
func setEnvForSpec(key, value string) {
	previous, had := os.LookupEnv(key)
	Expect(os.Setenv(key, value)).To(Succeed())
	DeferCleanup(func() {
		if had {
			Expect(os.Setenv(key, previous)).To(Succeed())
			return
		}
		Expect(os.Unsetenv(key)).To(Succeed())
	})
}

// isolateGitConfig detaches the spec from the developer's global and system git config, so an
// ambient setting (notably protocol.file.allow=always) cannot mask a missing production option.
func isolateGitConfig() {
	setEnvForSpec("GIT_CONFIG_GLOBAL", os.DevNull)
	setEnvForSpec("GIT_CONFIG_SYSTEM", os.DevNull)
	setEnvForSpec("GIT_TERMINAL_PROMPT", "0")
}

func gitAddSubmoduleSucceed(ctx context.Context, superRepoDir, url, path string) {
	gitSucceed(ctx, superRepoDir, "-c", "protocol.file.allow=always", "submodule", "add", url, path)
}

func gitUpdateSubmodulesSucceed(ctx context.Context, repoDir string, args ...string) {
	gitSucceed(ctx, repoDir, append([]string{"-c", "protocol.file.allow=always", "submodule", "update"}, args...)...)
}

func gitInitRepoWithFile(ctx context.Context, dir, fileName, content string) {
	gitInitRepo(ctx, dir)
	Expect(os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644)).To(Succeed())
	gitSucceed(ctx, dir, "add", ".")
	gitSucceed(ctx, dir, "commit", "-m", "content")
}
