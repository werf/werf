package releaseserver_test

import (
	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/integration/utils/liveexec"
)

func werfDeploy(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs(append([]string{"converge"}, extraArgs...)...)...)
}

func werfDismiss(dir string, opts liveexec.ExecCommandOptions) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs("dismiss", "--with-namespace")...)
}
