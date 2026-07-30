package true_git

import (
	"context"

	"github.com/werf/werf/v2/test/pkg/utils"
)

// gitCommitSucceed commits with signing disabled. Without that these tests inherit a
// developer's global commit.gpgsign and start depending on a working gpg-agent, which
// intermittently fails to sign under Ginkgo's parallel procs.
func gitCommitSucceed(ctx context.Context, workTreeDir string, args ...string) {
	utils.RunSucceedCommand(ctx, workTreeDir, "git", append([]string{"-c", "commit.gpgsign=false", "commit"}, args...)...)
}
