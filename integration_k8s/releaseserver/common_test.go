package releaseserver_test

import (
	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/liveexec"
)

func werfDeploy(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs(append([]string{"deploy", "--env", "dev"}, extraArgs...)...)...)
}

func werfDismiss(dir string, opts liveexec.ExecCommandOptions) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs("dismiss", "--env", "dev", "--with-namespace")...)
}
