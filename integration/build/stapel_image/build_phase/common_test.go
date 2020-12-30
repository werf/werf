package build_phase_test

import (
	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/integration/utils/liveexec"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"build"}, extraArgs...)...)...)
}

func werfPurge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"purge"}, extraArgs...)...)...)
}
