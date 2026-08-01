package true_git

import (
	"context"

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
