package ansible_test

import (
	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, append([]string{"build"}, extraArgs...)...)
}

func werfHostPurge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, append([]string{"host", "purge"}, extraArgs...)...)
}
