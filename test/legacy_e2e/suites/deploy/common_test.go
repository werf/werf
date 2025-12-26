package deploy_test

import (
	"context"

	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfConverge(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"converge"}, extraArgs...)...)
}
