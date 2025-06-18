package ansible_test

import (
	"context"

	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfBuild(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"build"}, extraArgs...)...)
}

func werfHostPurge(ctx context.Context, dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(ctx, dir, SuiteData.WerfBinPath, opts, append([]string{"host", "purge"}, extraArgs...)...)
}
